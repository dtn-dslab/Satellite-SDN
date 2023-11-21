package sdn

import (
	"fmt"
	"net/http"
	"time"

	"ws/dtn-satellite-sdn/sdn/clientset"
)

type HttpHandler func(http.ResponseWriter, *http.Request)

func RunSDNServer(url string, expectedNodeNum int, timeout int) error {
	// Create new clientset， set syncloop
	client := clientset.NewSDNClient(url)
	if err := client.ApplyTopo(); err != nil {
		return fmt.Errorf("apply topology error: %v", err)
	}
	if err := client.ApplyPod(expectedNodeNum); err != nil {
		return fmt.Errorf("apply pod error: %v", err)
	}
	if err := client.ApplyRoute(); err != nil {
		return fmt.Errorf("apply route error: %v", err)
	}
	if timeout != -1 {
		go func() {
			for {
				time.Sleep(time.Duration(timeout) * time.Second)
				if err := client.FetchAndUpdate(); err != nil {
					fmt.Printf("fetch and update topology error: %v\n", err)
				}
				if err := client.UpdateTopo(); err != nil {
					fmt.Printf("apply topology error: %v\n", err)
				}
				if err := client.UpdateRoute(); err != nil {
					fmt.Printf("apply route error: %v\n", err)
				}
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