# Docker Swarm Prometheus Exporter

Basic Prometheus exporter for a Docker Swarm.  

Exposes a `/metrics` HTTP endpoint for Prometheus to scrape which  has basic info on what services are running and how many tasks are in what state for each service.

## Running

There are two ways that you can run this:

### Standalone

You can run the binary on the docker host.  It will need permission to access the docker daemon.  

### Inside Docker

You can run the exporter as a container inside Docker.  For this to work you have to bind-mount `/var/run/docker.sock` from the host machine into the container.  An example docker-compose file is provided.

## Metrics Exposed

### `swarm_service_tasks`

Number of tasks for each service in each state.  At the moment we don't include zero values (which is against Prometheus guidelines but reduces the amount of data by an order of magnitude).

### `swarm_service_desired_replicas`

Number of tasks which each service is supposed to be running.

### `swarm_service_info`

Version info about the image used to run a service.  The useful stuff is in the labels, the value is always 1 (see [https://www.robustperception.io/exposing-the-software-version-to-prometheus](https://www.robustperception.io/exposing-the-software-version-to-prometheus)).

### `swarm_service_change_time`

The most recent time when any of the tasks for a service changed state, as a unix timestamp.


## Example

This is the (relevant) output from an example stack called `test` with two services - `redis` and the exporter itself.

```
# HELP swarm_service_change_time Time when a task state last changed
# TYPE swarm_service_change_time gauge
swarm_service_change_time{service_name="test_prometheus_exporter"} 1.581768126e+09
swarm_service_change_time{service_name="test_redis"} 1.581768125e+09

# HELP swarm_service_desired_replicas Number of replicas requested for this service
# TYPE swarm_service_desired_replicas gauge
swarm_service_desired_replicas{service_name="test_prometheus_exporter"} 1
swarm_service_desired_replicas{service_name="test_redis"} 1

# HELP swarm_service_info Information about each service
# TYPE swarm_service_info gauge
swarm_service_info{image="docker_swarm_exporter:latest",service_name="test_prometheus_exporter"} 1
swarm_service_info{image="redis:latest@sha256:7b84b346c01e5a8d204a5bb30d4521bcc3a8535bbf90c660b8595fad248eae82",service_name="test_redis"} 1

# HELP swarm_service_tasks Number of docker tasks
# TYPE swarm_service_tasks gauge
swarm_service_tasks{service_name="test_prometheus_exporter",state="running"} 1
swarm_service_tasks{service_name="test_redis",state="preparing"} 1
swarm_service_tasks{service_name="test_redis",state="running"} 0
```