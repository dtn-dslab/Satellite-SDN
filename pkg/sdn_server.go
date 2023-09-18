package sdn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path"

	"ws/dtn-satellite-sdn/pkg/link"
	"ws/dtn-satellite-sdn/pkg/partition"
	"ws/dtn-satellite-sdn/pkg/route"
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

// Construct network emulation system with nameMap, edgeSet, routeTable and expectedNodeNum
func CreateSDN(nameMap map[int]string, edgeSet []link.LinkEdge, routeTable [][]int, expectedNodeNum int) error {
	// Define the path that yaml files store in
	podOutputPath := path.Join("./pkg/output", "pod.yaml")
	topoOutputPath := path.Join("./pkg/output", "topology.yaml")
	routeOutputPath := path.Join("./pkg/output", "route.yaml")

	// Generate topology file & apply topology
	log.Println("Generating topology yaml...")
	err := link.GenerateLinkSummaryFile(nameMap, edgeSet, topoOutputPath)
	if err != nil {
		return fmt.Errorf("Generating topology yaml failed: %v\n", err)
	}
	topoCmd := exec.Command("kubectl", "apply", "-f", topoOutputPath)
	if err = topoCmd.Run(); err != nil {
		return fmt.Errorf("Apply topology error: %v\n", err)
	}

	// Generate pod file & apply pod
	// p.s. We need to apply topology first due to the implementation of kube-dtn.
	log.Println("Generate pod yaml...")
	err = partition.GeneratePodSummaryFile(nameMap, edgeSet, podOutputPath, expectedNodeNum)
	if err != nil {
		return fmt.Errorf("Generating pod yaml failed: %v\n", err)
	}
	podCmd := exec.Command("kubectl", "apply", "-f", podOutputPath)
	if err = podCmd.Run(); err != nil {
		return fmt.Errorf("Apply pod error: %v\n", err)
	}

	// Generate route file & apply route
	log.Println("Generating route yaml...")
	err = route.GenerateRouteSummaryFile(nameMap, routeTable, routeOutputPath)
	if err != nil {
		return fmt.Errorf("Generating route yaml failed: %v\n", err)
	}
	routeCmd := exec.Command("kubectl", "apply", "-f", routeOutputPath)
	if err = routeCmd.Run(); err != nil {
		return fmt.Errorf("Apply route error: %v\n", err)
	}

	return nil
}

// Update network emulation system with nameMap, edgeSet and routeTable
func UpdateSDN(nameMap map[int]string, edgeSet []link.LinkEdge, routeTable [][]int) error {
	// Define the path that yaml files store in
	topoOutputPath := path.Join("./output", "topology.yaml")
	routeOutputPath := path.Join("./output", "route.yaml")

	// Generate topology file & apply topology
	log.Println("Generating topology yaml...")
	err := link.GenerateLinkSummaryFile(nameMap, edgeSet, topoOutputPath)
	if err != nil {
		return fmt.Errorf("Generating topology yaml failed: %v\n", err)
	}
	topoCmd := exec.Command("kubectl", "apply", "-f", topoOutputPath)
	if err = topoCmd.Run(); err != nil {
		return fmt.Errorf("Apply topology error: %v\n", err)
	}

	// Generate route file & apply route
	log.Println("Generating route yaml...")
	err = route.GenerateRouteSummaryFile(nameMap, routeTable, routeOutputPath)
	if err != nil {
		return fmt.Errorf("Generating route yaml failed: %v\n", err)
	}
	routeCmd := exec.Command("kubectl", "apply", "-f", routeOutputPath)
	if err = routeCmd.Run(); err != nil {
		return fmt.Errorf("Apply route error: %v\n", err)
	}

	return nil
}

// Uninitialize the network emulation system
func DelSDN(nameMap map[int]string) error {
	// Define the path that yaml files store in
	podOutputPath := path.Join("./output", "pod.yaml")
	topoOutputPath := path.Join("./output", "topology.yaml")
	routeOutputPath := path.Join("./output", "route.yaml")

	// Delete Pod
	podCmd := exec.Command("kubectl", "delete", "-f", podOutputPath)
	if err := podCmd.Run(); err != nil {
		return fmt.Errorf("Delete pod error: %v\n", err)
	}
	// Delete Topology
	topoCmd := exec.Command("kubectl", "delete", "-f", topoOutputPath)
	if err := topoCmd.Run(); err != nil {
		return fmt.Errorf("Delete topology error: %v\n", err)
	}
	// Delete Route
	routeCmd := exec.Command("kubectl", "delete", "-f", routeOutputPath)
	if err := routeCmd.Run(); err != nil {
		return fmt.Errorf("Delete route error: %v\n", err)
	}

	return nil
}

func Run() {
	// Bind http request with handler
	for url, handler := range sdnHandlerMap {
		http.HandleFunc(url, handler)
	}

	// Start Server
	http.ListenAndServe(":30101", nil)
}