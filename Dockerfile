FROM alpine
COPY ./docker_swarm_exporter /usr/local/bin/docker_swarm_exporter
CMD [ "docker_swarm_exporter" ]
EXPOSE 9675/tcp
