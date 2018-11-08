#!/usr/bin/env bash

[ -z "$HOSTMAP" ] && echo "No Host Map Configuration (env HOSTMAP)" && exit 1

echo -e "
user              nginx;
worker_processes  auto;
error_log         /var/log/nginx/error.log warn;
pid               /var/run/nginx.pid;

events {
    worker_connections  1024;
}

stream {
    server {
        listen 443;
        proxy_pass \$name;
        ssl_preread on;
    }

    map \$ssl_preread_server_name \$name {" > /etc/nginx/nginx.conf

declare -A HOSTMAPARRAY
IFS=',' read -r -a array <<< "$HOSTMAP"
for elememt in "${array[@]}"
do
    HOSTMAPARRAY[$(echo ${elememt} | cut -d' ' -f 1)]=$(echo ${elememt} | cut -d' ' -f 2)
done

for Key in "${!HOSTMAPARRAY[@]}"
do
    upstreamName=$(echo ${HOSTMAPARRAY[$Key]} | cut -d':' -f 1)
    if [[ ! -z "$(nc ${upstreamName} 2>&1 | grep 'forward host lookup failed')" ]]
    then
        unset HOSTMAPARRAY[$Key]
    fi
done

for Key in "${!HOSTMAPARRAY[@]}"
do
    upstreamName=$(echo ${HOSTMAPARRAY[$Key]} | cut -d':' -f 1)$(echo ${HOSTMAPARRAY[$Key]} | cut -d':' -f 2)
    echo -e "      ${Key} ${upstreamName};" >> /etc/nginx/nginx.conf
done

echo -e "    }
" >> /etc/nginx/nginx.conf

for Key in "${!HOSTMAPARRAY[@]}"
do
    upstreamName=$(echo ${HOSTMAPARRAY[$Key]} | cut -d':' -f 1)$(echo ${HOSTMAPARRAY[$Key]} | cut -d':' -f 2)
    echo -e "    upstream ${upstreamName} {
        server ${HOSTMAPARRAY[$Key]};
    }
" >> /etc/nginx/nginx.conf
done

echo -e "}
" >> /etc/nginx/nginx.conf

cat /etc/nginx/nginx.conf

if [[ ! -z "$(service nginx status 2>&1 | grep 'nginx is running')" ]]
then
    service nginx reload
else
    nginx -g "daemon off;"
fi



