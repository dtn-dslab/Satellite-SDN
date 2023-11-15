package sdn

import (
	"net/http"
	"time"

	"ws/dtn-satellite-sdn/sdn/clientset"
)

type HttpHandler func(http.ResponseWriter, *http.Request)

func RunSDNServer(url string, expectedNodeNum int, timeout int) error {
	// Create new clientset， set syncloop
	client := clientset.NewSDNClient(url)
	client.ApplyTopo()
	client.ApplyPod(expectedNodeNum)
	client.ApplyRoute()
	if timeout != -1 {
		go func() {
			for {
				time.Sleep(time.Duration(timeout) * time.Second)
				client.FetchAndUpdate()
				client.ApplyTopo()
				client.ApplyRoute()
			}
		}()
	}

	// Bind http request with handler
	sdnHandlerMap := map[string]HttpHandler {
		"/getTopologyGraph": client.GetTopoInAscArrayHandler,
		"/getRoute": client.GetRouteFromAndToHandler,
		"/getConnection": client.GetRouteHopsHandler,
	}
	for url, handler := range sdnHandlerMap {
		http.HandleFunc(url, handler)
	}

	// Start Server
	return http.ListenAndServe(":30101", nil)
}