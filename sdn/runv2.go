package sdn

import (
	"fmt"
	"log"
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
	log.Println("Done!")
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
				log.Println("Done!")
			}
		}()
	}

	// Bind http request with handler
	sdnHandlerMap := map[string]HttpHandler{
		"/getTopologyGraph": client.GetTopoInAscArrayHandler,
		"/getRoute":         client.GetRouteFromAndToHandler,
		"/getConnection":    client.GetRouteHopsHandler,
		"/getDistance":      client.GetDistanceHanlder,
	}
	for url, handler := range sdnHandlerMap {
		http.HandleFunc(url, handler)
	}

	// Start Server
	return http.ListenAndServe(":30101", nil)
}

func RunSDNServerTest(url string, expectedNodeNum int, timeout int) error {
	// Create new clientset， set syncloop
	client := clientset.NewSDNClient(url)
	log.Println("Done!")
	if timeout != -1 {
		go func() {
			for {
				time.Sleep(time.Duration(timeout) * time.Second)
				if err := client.FetchAndUpdate(); err != nil {
					fmt.Printf("fetch and update topology error: %v\n", err)
				}
				log.Println("Done!")
			}
		}()
	}

	// Bind http request with handler
	sdnHandlerMap := map[string]HttpHandler{
		"/getTopologyGraph": client.GetTopoInAscArrayHandler,
		"/getRoute":         client.GetRouteFromAndToHandler,
		"/getConnection":    client.GetRouteHopsHandler,
		"/getDistance":      client.GetDistanceHanlder,
	}
	for url, handler := range sdnHandlerMap {
		http.HandleFunc(url, handler)
	}

	// Start Server
	return http.ListenAndServe(":30102", nil)
}
