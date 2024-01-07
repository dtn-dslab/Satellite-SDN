package sdn

import (
	"log"
	"net/http"
	"os/exec"
	"time"

	"ws/dtn-satellite-sdn/sdn/clientset"

	"github.com/sirupsen/logrus"
)

type HttpHandler func(http.ResponseWriter, *http.Request)

func RunSDNServer(url string, expectedNodeNum int, timeout int) error {
	// Create new clientset
	logger := logrus.WithFields(logrus.Fields{
		"url": 		url,
		"node-num": expectedNodeNum,
		"timeout":  timeout,
	})
	logger.WithField("time", time.Now()).Info("start sdn server")

	client := clientset.NewSDNClient(url)
	if err := client.ApplyTopo(); err != nil {
		logger.WithError(err).Error("apply topology failed")
		return err
	}
	if err := client.ApplyPod(expectedNodeNum); err != nil {
		logger.WithError(err).Error("apply pod failed")
		return err
	}
	if err := client.ApplyRoute(); err != nil {
		logger.WithError(err).Error("apply route failed")
		return err
	}
	logger.WithField("time", time.Now()).Info("sdn server has been started!")

	// Set up sync loop
	if timeout != -1 {
		go func() {
			for {
				time.Sleep(time.Duration(timeout) * time.Second)
				logger.WithField("time", time.Now()).Info("update sdn server.")
				if err := client.FetchAndUpdate(); err != nil {
					logger.WithError(err).Error("fetch and update topology err")
				}
				if err := client.UpdateTopo(); err != nil {
					logger.WithError(err).Error("update topology error")
				}
				if err := client.UpdateRoute(); err != nil {
					logger.WithError(err).Error("update route error")
				}
				logger.WithField("time", time.Now()).Info("sdn server has been updated!")
			}
		}()
	}

	// Bind http request with handler
	sdnHandlerMap := map[string]HttpHandler{
		"/getTopologyGraph": client.GetTopoInAscArrayHandler,
		"/getRoute":         client.GetRouteFromAndToHandler,
		"/getConnection":    client.GetRouteHopsHandler,
		"/getDistance":      client.GetDistanceHanlder,
		"/getSpreadArray":	 client.GetSpreadArrayHanlder,
	}
	for url, handler := range sdnHandlerMap {
		http.HandleFunc(url, handler)
	}

	// Start Server
	return http.ListenAndServe(":30101", nil)
}

func RunSDNServerTest(url string, expectedNodeNum int, timeout int) error {
	// Create new clientset
	logger := logrus.WithFields(logrus.Fields{
		"url": 		url,
		"node-num": expectedNodeNum,
		"timeout":  timeout,
	})
	logger.WithField("time", time.Now()).Info("start sdn server.")

	client := clientset.NewSDNClient(url)
	logger.WithField("time", time.Now()).Info("sdn server has been started!")

	// Set up sync loop
	if timeout != -1 {
		go func() {
			for {
				time.Sleep(time.Duration(timeout) * time.Second)
				logger.WithField("time", time.Now()).Info("update sdn server.")
				if err := client.FetchAndUpdate(); err != nil {
					logger.WithError(err).Error("fetch and update topology error")
				}
				logger.WithField("time", time.Now()).Info("sdn server has been updated!")
			}
		}()
	}

	// Bind http request with handler
	sdnHandlerMap := map[string]HttpHandler{
		"/getTopologyGraph": client.GetTopoInAscArrayHandler,
		"/getRoute":         client.GetRouteFromAndToHandler,
		"/getConnection":    client.GetRouteHopsHandler,
		"/getDistance":      client.GetDistanceHanlder,
		"/getSpreadArray":	 client.GetSpreadArrayHanlder,
		"/metrics":			 client.GetFakeMetricsHandler,
	}
	for url, handler := range sdnHandlerMap {
		http.HandleFunc(url, handler)
	}

	// Start Server
	return http.ListenAndServe(":30102", nil)
}

func RunRestartTestServer(scriptPath string) error {
	http.HandleFunc("/restart", func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command(scriptPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("[Error]: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(output)
		} else {
			log.Println("Restart successfully!")
			w.WriteHeader(http.StatusOK)
		}
	})
	return http.ListenAndServe(":30103", nil)
}
