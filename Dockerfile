FROM nginx:1.15.2

RUN apt-get update
RUN apt-get install -y netcat procps

COPY bin/init.sh /init.sh
RUN chmod +x /init.sh

WORKDIR /etc/nginx

ENTRYPOINT ["/bin/bash"]
CMD ["/init.sh"]
