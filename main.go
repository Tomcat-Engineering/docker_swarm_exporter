/*
Basic Prometheus exporter for a Docker Swarm.  Exposes a HTTP endpoint for
Prometheus to scrape which just has the basic info on what services are
running and how many tasks are in what state for each service.
*/
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	coll := DockerServices{Client: dockerClient}
	if err := prometheus.Register(&coll); err != nil {
		panic(err)
	}

	// Get rid of the stupid golang metrics
	prometheus.Unregister(collectors.NewGoCollector())

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":9675", nil); err != nil {
		panic(err)
	}
}

// DockerServices implements the Collector interface.
type DockerServices struct {
	*client.Client
}

var _ prometheus.Collector = (*DockerServices)(nil)

var (
	replicaCount = prometheus.NewDesc(
		"swarm_service_desired_replicas",
		"Number of replicas requested for this service",
		[]string{"service_name"}, nil,
	)
	taskCount = prometheus.NewDesc(
		"swarm_service_tasks",
		"Number of docker tasks",
		[]string{"service_name", "state"}, nil,
	)
	imageVersion = prometheus.NewDesc(
		"swarm_service_info",
		"Information about each service",
		[]string{"service_name", "image"}, nil,
	)
	lastChangeTime = prometheus.NewDesc(
		"swarm_service_change_time",
		"Time when a task state last changed",
		[]string{"service_name"}, nil,
	)
)

func (c DockerServices) Describe(ch chan<- *prometheus.Desc) {
	ch <- replicaCount
	ch <- taskCount
	ch <- imageVersion
	ch <- lastChangeTime
}

// Collect scrapes the container information from Docker.
func (c DockerServices) Collect(ch chan<- prometheus.Metric) {

	services, err := c.Client.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		panic(err)
	}

	tasks, err := c.Client.TaskList(context.Background(), types.TaskListOptions{})
	if err != nil {
		panic(err)
	}

	for _, service := range services {
		if service.Spec.Mode.Replicated != nil {
			ch <- prometheus.MustNewConstMetric(
				replicaCount,
				prometheus.GaugeValue,
				float64(*service.Spec.Mode.Replicated.Replicas),
				service.Spec.Annotations.Name,
			)
		}

		taskStates := make(map[string]int)
		taskStates["running"] = 0 // Should really do this for all potential states (https://github.com/moby/moby/blob/v1.13.1/api/types/swarm/task.go)
		var lastTaskStatusChange time.Time
		for _, task := range tasks {
			if task.ServiceID == service.ID {
				taskStates[string(task.Status.State)] += 1
				if task.Status.Timestamp.After(lastTaskStatusChange) {
					lastTaskStatusChange = task.Status.Timestamp
				}
			}
		}

		for state, count := range taskStates {
			ch <- prometheus.MustNewConstMetric(
				taskCount,
				prometheus.GaugeValue,
				float64(count),
				service.Spec.Annotations.Name,
				string(state),
			)
		}

		// See https://www.robustperception.io/exposing-the-software-version-to-prometheus
		ch <- prometheus.MustNewConstMetric(
			imageVersion,
			prometheus.GaugeValue,
			1,
			service.Spec.Annotations.Name,
			string(service.Spec.TaskTemplate.ContainerSpec.Image),
		)

		ch <- prometheus.MustNewConstMetric(
			lastChangeTime,
			prometheus.GaugeValue,
			float64(lastTaskStatusChange.Unix()),
			service.Spec.Annotations.Name,
		)
	}
}
