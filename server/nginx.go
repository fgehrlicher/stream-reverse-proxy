package server

import (
	"sync"
	"path/filepath"
	"os"
	"os/exec"
	"errors"
	"fmt"
	"strconv"
	"io/ioutil"
	"crypto/md5"
	"encoding/hex"
)

var (
	InvalidServiceError = errors.New("invalid Service")
)

const NginxConfigFilePath = "/etc/nginx/nginx.conf"

const NginxGeneralConf = `
user              nginx;
worker_processes  auto;
error_log         /var/log/nginx/error.log warn;
pid               /var/run/nginx.pid;
`

const NginxEventConf = `
events {
    worker_connections  1024;
}
`

const NginxStream = `
stream {

    server {
        listen 443;
        proxy_pass $name;
        ssl_preread on;
    }
`

const NginxStreamEnd = "}"

const NginxMap = `
    map $ssl_preread_server_name $name {
`

const NginxMapEnd = "    }\n\n"

func startNginx() error {
	cmd := exec.Command("/bin/sh", "-c", "nginx")
	err := cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		err = cmd.Wait()
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func reloadNginx() error {
	_, err := exec.Command("/bin/sh", "-c", "service nginx reload").Output()
	if err == nil {
		fmt.Println("Reloaded Nginx")
	}
	return err
}

type NginxConfig struct {
	sync     sync.RWMutex
	services []Service
}

func (this *NginxConfig) AddService(nginxService Service) error {
	this.sync.Lock()
	found := false
	for index, value := range this.services {
		if value.SubDomain == nginxService.SubDomain {
			this.services[index] = nginxService
			found = true
		}
	}
	if !found {
		this.services = append(this.services, nginxService)
	}
	this.sync.Unlock()
	return nil
}

func (this *NginxConfig) WriteConfig() error {
	this.sync.Lock()
	configDir := filepath.Dir(NginxConfigFilePath)
	_, err := os.Stat(configDir)
	if err != nil {
		return err
	}
	config, err := this.renderConfig()
	if err != nil {
		return err
	}
	configByte := []byte(config)
	err = ioutil.WriteFile(NginxConfigFilePath, configByte, 0644)
	if err == nil {
		fmt.Println("ReWritten Config:")
		fmt.Println(config + "\n")
		err = reloadNginx()
	}
	this.sync.Unlock()
	return err
}

func (this *NginxConfig) renderConfig() (string, error) {
	config := NginxGeneralConf + NginxEventConf + NginxStream + NginxMap

	for _, service := range this.services {
		serviceMap, err := service.ConvertToMap()
		if err != nil {
			return "", err
		}
		config = config + serviceMap
	}
	config = config + NginxMapEnd

	for _, service := range this.services {
		serviceUpstream, err := service.ConvertToUpstream()
		if err != nil {
			return "", err
		}
		config = config + serviceUpstream
	}
	config = config + NginxStreamEnd
	return config, nil
}

type Service struct {
	SubDomain string `json:"sub_domain"`
	Ip        string `json:"ip"`
	Port      int    `json:"port"`
}

func (this *Service) getHash() string {
	hasher := md5.New()
	hasher.Write([]byte(this.Ip + strconv.Itoa(this.Port) + this.SubDomain))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (this *Service) ConvertToMap() (string, error) {
	if this.Ip == "" || this.SubDomain == "" || this.Port == 0 {
		return "", InvalidServiceError
	}

	mapString := fmt.Sprintf("      %s %s;\n", this.SubDomain, this.getHash())
	return mapString, nil
}

func (this *Service) ConvertToUpstream() (string, error) {
	if this.Ip == "" || this.SubDomain == "" || this.Port == 0 {
		return "", InvalidServiceError
	}
	upstreamString := fmt.Sprintf("    upstream %s {\n", this.getHash()) +
		fmt.Sprintf("        server %s:%d;\n", this.Ip, this.Port) + "    }\n\n"
	return upstreamString, nil
}
