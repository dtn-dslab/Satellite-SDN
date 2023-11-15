package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	log "github.com/sirupsen/logrus"
	"github.com/y-young/kube-dtn/common"
	"github.com/y-young/kube-dtn/daemon/vxlan"
	pb "github.com/y-young/kube-dtn/proto/v1"
)

type Query struct {
	Link     pb.Link `json:"link"`
	RemoteIP string  `json:"remote_ip"`
}

func (q *Query) Print() {
	log.Infof("Link: %s", q.Link.String())
	log.Infof("RemoteIP: %s", q.RemoteIP)
}

func main() {
	intf := flag.String("i", "", "Interface for vxlan")
	ip := flag.String("a", "", "Local IP for vxlan")
	path := flag.String("f", "", "Path to yaml file")
	flag.Parse()
	if *path == "" {
		log.Fatalf("Please provide a path to a yaml file")
		return
	}

	file, err := os.ReadFile(*path)
	if err != nil {
		log.Fatalf("Failed to read configuration: %s", err)
		return
	}

	var query Query
	err = yaml.Unmarshal(file, &query)
	if err != nil {
		log.Fatalf("Failed to parse configuration: %s", err)
		return
	}
	link := &query.Link
	query.Print()

	pseudoPod := link.PeerPod
	if strings.HasPrefix(pseudoPod, "physical/") {
		*ip = strings.TrimPrefix(pseudoPod, "physical/")
	} else {
		log.Fatalf("Peer pod name should begin with \"physical/\": %s", pseudoPod)
		return
	}

	if *intf == "" && *ip == "" {
		*intf, *ip, err = vxlan.GetDefaultVxlanSource()
		if err != nil {
			log.Fatalf("Failed to get default vxlan source: %s", err)
			return
		}
	} else if *intf == "" {
		*intf, _, err = vxlan.GetVxlanSource(*ip)
		if err != nil {
			log.Fatalf("Failed to get vxlan source with IP %s: %s", *ip, err)
			return
		}
	}
	if *intf == "" || *ip == "" {
		log.Fatalf("Failed to get vxlan source, please specify manually")
		return
	}

	err = addLink(link, *intf, query.RemoteIP)
	if err != nil {
		log.Errorf("Failed to add link: %s", err)
	} else {
		log.Info("Successfully added link locally, now add this link to topology CRD")
	}
}

func addLink(link *pb.Link, srcIntf string, peerVtep string) error {
	// We're connecting physical host interface, so use root network namespace
	vxlanSpec := &vxlan.VxlanSpec{
		NetNs: "",
		// Link in configuration file is from pod's perspective, so we need to reverse it
		IntfName: link.PeerIntf,
		IntfIp:   link.PeerIp,
		PeerVtep: peerVtep,
		Vni:      common.GetVniFromUid(link.Uid),
	}
	ctx := context.WithValue(context.Background(), common.CtxKey("logger"), log.New())
	if err := vxlan.SetupVxLan(ctx, vxlanSpec, link.Properties); err != nil {
		log.Infof("Error when creating a Vxlan interface with koko: %s", err)
		return err
	}
	return nil
}
