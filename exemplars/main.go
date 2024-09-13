package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Stream struct {
	Stream struct {
		Traces string `json:"traces"`
	} `json:"stream"`
	Values [][]string `json:"values"`
}

type Streams struct {
	Streams []Stream `json:"streams"`
}

func connectWebsocket(urlString string) *websocket.Conn {
	log.Print("connecting to websocket")

	c, _, err := websocket.DefaultDialer.Dial(urlString, http.Header{
		"X-Scope-OrgId": []string{"tenant1"},
	})
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Print("connected to websocket")

	return c
}

func main() {
	requestDurations := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "prometheus_exemplar",
		Help:    "A histogram of the HTTP request durations in seconds.",
		Buckets: prometheus.ExponentialBuckets(0.1, 1.5, 5),
	})
	// Create non-global registry.
	registry := prometheus.NewRegistry()

	// Add go runtime metrics and process collectors.
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		requestDurations,
	)

	host, exists := os.LookupEnv("HOST")
	if !exists || host == "" {
		log.Fatal("No HOST defined, where am I pulling trace ID's from?")
	}
	log.Printf("successfully retrieved host %s", host)

	port, exists := os.LookupEnv("PORT")
	if !exists || port == "" {
		log.Fatal("No PORT defined, where am I pulling trace ID's from?")
	}
	log.Printf("successfully retrieved port %s", port)

	endpoint := "loki/api/v1/tail"
	urlString := fmt.Sprintf("ws://%s:%s/%s?query=%%7Btraces%%3D%%7E%%22.%%2B%%22%%7D&limit=1000", host, port, endpoint)
	log.Printf("url: %s", urlString)

	time.Sleep(15 * time.Second)

	c := connectWebsocket(urlString)
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			streams := Streams{}
			err := c.ReadJSON(&streams)
			if err != nil {
				log.Println("read:", err)
				if strings.Contains(err.Error(), "reached tail max duration limit") {
					c.Close()
					c = connectWebsocket(urlString)
				} else {
					return
				}
			}
			for _, trace := range streams.Streams {
				log.Print("parsing trace from stream")
				splitString := strings.Split(trace.Values[0][1], " ")
				traceID := ""
				for _, kv := range splitString {
					split := strings.Split(kv, "=")
					if len(split) != 0 && split[0] == "tid" {
						traceID = split[1]
						log.Println("traceID:", traceID)
					}
				}
				now := time.Now()
				requestDurations.(prometheus.ExemplarObserver).ObserveWithExemplar(
					time.Since(now).Seconds(), prometheus.Labels{"traceID": traceID},
				)
				log.Printf("pushed metric for traceID: %s", traceID)
			}
		}
	}()

	// Expose /metrics HTTP endpoint using the created custom registry.
	http.Handle(
		"/metrics", promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			}),
	)
	// To test: curl -H 'Accept: application/openmetrics-text' localhost:8081/metrics
	log.Fatalln(http.ListenAndServe(":8081", nil))
}
