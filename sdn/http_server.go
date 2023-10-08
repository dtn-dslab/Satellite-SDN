package sdn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"ws/dtn-satellite-sdn/sdn/link"
)

type HttpHandler func(http.ResponseWriter, *http.Request)

type SDNParam struct {
	NameMap map[int]string `json:"nameMap,omitempty"`
	EdgeSet []link.LinkEdge `json:"edgeSet,omitempty"`
	RouteTable [][]int `json:"routeTable,omitempty"`
	ExpectedNodeNum int `json:"expectedNodeNum,omitempty"`
}

var sdnHandlerMap = map[string]HttpHandler {
	"/sdn/create": CreateSDNHandler,
	"/sdn/update": UpdateSDNHandler,
	"/sdn/del": DelSDNHandler,
}

func CreateSDNHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	// Parsing params
	param := SDNParam{}
	json.Unmarshal(body, &param)
	if param.NameMap == nil || param.EdgeSet == nil || param.RouteTable == nil ||
		len(param.NameMap) == 0 || len(param.EdgeSet) == 0 || len(param.RouteTable) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid arguments!"))
		return
	}

	// Create SDN environment
	err := CreateSDN(param.NameMap, param.EdgeSet, param.RouteTable, param.ExpectedNodeNum)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("CreateSDN error: %v\n", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Create SDN successfully!"))
}

func UpdateSDNHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	// Parsing params
	param := SDNParam{}
	json.Unmarshal(body, &param)
	if param.NameMap == nil || param.EdgeSet == nil || param.RouteTable == nil ||
		len(param.NameMap) == 0 || len(param.EdgeSet) == 0 || len(param.RouteTable) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid arguments!"))
		return
	}

	// Update SDN environment
	err := UpdateSDN(param.NameMap, param.EdgeSet, param.RouteTable)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("UpdateSDN error: %v\n", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Update SDN successfully!"))
}

func DelSDNHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	// Parsing params
	param := SDNParam{}
	json.Unmarshal(body, &param)
	if param.NameMap == nil || len(param.NameMap) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid arguments!"))
		return
	}

	// Delete SDN environment
	if err := DelSDN(param.NameMap); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("DelSDN error: %v\n", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Delete SDN successfully!"))
}

func Run() {
	// Bind http request with handler
	for url, handler := range sdnHandlerMap {
		http.HandleFunc(url, handler)
	}

	// Start Server
	http.ListenAndServe(":30101", nil)
}