package server

import (
	"net/http"
	"log"
	"fmt"
	"net"
	"strconv"
)

var NginxConf = &NginxConfig{}

func StartListener() {
	err := startNginx()
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", registrationEndpoint)
	fmt.Println("Server started on port 5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func registrationEndpoint(response http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		invalidMethodEndpoint(response)
		return
	}
	ip, _, _ := net.SplitHostPort(request.RemoteAddr)
	subdomain := request.Header.Get("subdomain")
	port := request.Header.Get("port")
	portInt, err := strconv.Atoi(port)

	if err != nil {
		fmt.Println(err.Error())
		invalidMethodEndpoint(response)
		return
	}

	service := Service{
		Ip:        ip,
		SubDomain: subdomain,
		Port:      portInt,
	}

	err = NginxConf.AddService(service)
	if err != nil {
		fmt.Println(err.Error())
		ServerErrorEndpoint(response)
		return
	}

	err = NginxConf.WriteConfig()
	if err != nil {
		fmt.Println(err.Error())
		ServerErrorEndpoint(response)
		return
	}

	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	response.Header().Set("X-Content-Type-Options", "nosniff")
	response.WriteHeader(200)
	fmt.Fprintln(response, "Added Subdomain.")
}
