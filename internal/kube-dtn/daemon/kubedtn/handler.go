package kubedtn

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	v1 "github.com/y-young/kube-dtn/api/v1"
	"github.com/y-young/kube-dtn/daemon/grpcwire"
	"github.com/y-young/kube-dtn/daemon/vxlan"

	"github.com/davecgh/go-spew/spew"
	koko "github.com/redhat-nfvpe/koko/api"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/google/gopacket/pcap"
	"github.com/y-young/kube-dtn/common"
	pb "github.com/y-young/kube-dtn/proto/v1"
)

// var interNodeLinkType = common.INTER_NODE_LINK_VXLAN

func (m *KubeDTN) getPod(ctx context.Context, name, ns string) (*v1.Topology, error) {
	logger := common.GetLogger(ctx)
	if ns == "" {
		ns = "default"
	}
	pod := fmt.Sprintf("%s/%s", ns, name)
	logger.Infof("Reading pod %s from informer", pod)

	obj, exists, err := m.topologyStore.GetByKey(pod)
	if err != nil || !exists {
		logger.Infof("Failed to read pod %s from informer, trying from K8s", name)
		return m.tClient.Topology(ns).Get(ctx, name, metav1.GetOptions{})
	}
	return obj.(*v1.Topology), nil
}

func (m *KubeDTN) updateStatus(ctx context.Context, topology *v1.Topology, ns string) error {
	logger := common.GetLogger(ctx)
	logger.Infof("Update pod status %s from K8s", topology.Name)
	_, err := m.tClient.Topology(ns).UpdateStatus(ctx, topology, metav1.UpdateOptions{})
	return err
}

func (m *KubeDTN) Get(ctx context.Context, pod *pb.PodQuery) (*pb.Pod, error) {
	logger := common.GetLogger(ctx)

	topology, err := m.getPod(ctx, pod.Name, pod.KubeNs)
	if err != nil {
		logger.Errorf("Failed to read pod %s", pod.Name)
		return nil, err
	}

	return m.ToProtoPod(ctx, topology)
}

func (m *KubeDTN) ToProtoPod(ctx context.Context, topology *v1.Topology) (*pb.Pod, error) {
	logger := common.GetLogger(ctx)

	remoteLinks := topology.Spec.Links
	if remoteLinks == nil {
		logger.Errorf("Could not find 'Link' array in pod's spec")
		return nil, fmt.Errorf("could not find 'Link' array in pod's spec")
	}

	links := make([]*pb.Link, len(remoteLinks))
	for i := range links {
		remoteLink := remoteLinks[i]
		newLink := remoteLink.ToProto()
		links[i] = newLink
	}

	srcIP := topology.Status.SrcIP
	netNs := topology.Status.NetNs

	return &pb.Pod{
		Name:   topology.Name,
		SrcIp:  srcIP,
		NetNs:  netNs,
		KubeNs: topology.Namespace,
		Links:  links,
	}, nil
}

func (m *KubeDTN) SetAlive(ctx context.Context, pod *pb.Pod) (*pb.BoolResponse, error) {
	logger := common.GetLogger(ctx).WithFields(log.Fields{
		"pod":    pod.Name,
		"ns":     pod.KubeNs,
		"action": "setAlive",
	})
	ctx = common.WithLogger(ctx, logger)

	logger.Infof("Setting SrcIp=%s and NetNs=%s", pod.SrcIp, pod.NetNs)
	alive := pod.SrcIp != "" && pod.NetNs != ""

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		topology, err := m.getPod(ctx, pod.Name, pod.KubeNs)
		if err != nil {
			logger.Errorf("Failed to read pod from K8s: %v", err)
			return err
		}

		if alive {
			m.topologyManager.Add(topology)
		} else {
			m.topologyManager.Delete(topology.Name, topology.Namespace)
		}
		topology.Status.SrcIP = pod.SrcIp
		topology.Status.NetNs = pod.NetNs

		return m.updateStatus(ctx, topology, pod.KubeNs)
	})

	if retryErr != nil {
		logger.Errorf("Failed to update pod alive status: %v", retryErr)
		return &pb.BoolResponse{Response: false}, retryErr
	}

	// UpdateStatus can only update the status field, but we need to update metadata
	retryErr = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		topology, err := m.getPod(ctx, pod.Name, pod.KubeNs)
		if err != nil {
			logger.Errorf("Failed to read pod from K8s: %v", err)
			return err
		}

		if alive {
			topology.Finalizers = append(topology.Finalizers, v1.GroupVersion.Identifier())
		} else {
			topology.Finalizers = []string{}
		}

		_, err = m.tClient.Topology(pod.KubeNs).Update(ctx, topology, metav1.UpdateOptions{})
		return err
	})

	if retryErr != nil {
		logger.Errorf("Failed to set finalizer: %v", retryErr)
	}

	return &pb.BoolResponse{Response: true}, nil
}

func (m *KubeDTN) Update(ctx context.Context, pod *pb.RemotePod) (*pb.BoolResponse, error) {
	uid := common.GetUidFromVni(pod.Vni)
	logger := common.GetLogger(ctx).WithFields(log.Fields{
		"pod":    pod.Name,
		"ns":     pod.KubeNs,
		"link":   uid,
		"action": "remoteUpdate",
	})
	ctx = common.WithLogger(ctx, logger)
	logger.Infof("Updating pod from remote")
	startTime := time.Now()

	var err error
	vxlanSpec := &vxlan.VxlanSpec{
		NetNs:    pod.NetNs,
		IntfName: pod.IntfName,
		IntfIp:   pod.IntfIp,
		PeerVtep: pod.PeerVtep,
		Vni:      pod.Vni,
		SrcIp:    m.nodeIP,
		SrcIntf:  m.vxlanIntf,
	}

	mutex := m.linkMutexes.Get(uid)
	mutex.Lock()
	defer mutex.Unlock()

	// Check if there's a vxlan link with the same VNI but in different namespace
	netns := m.vxlanManager.Get(pod.Vni)
	// Ensure link is in different namespace since we might have set it up locally
	if netns != nil && *netns != pod.NetNs {
		logger.Infof("VXLAN with the same VNI already exists, removing it")
		err = vxlan.RemoveLinkWithVni(ctx, pod.Vni, *netns)
		if err != nil {
			logger.Errorf("Failed to remove existing VXLAN link: %v", err)
		}
	}

	err = vxlan.SetupVxLan(ctx, vxlanSpec, pod.Properties)
	if err != nil {
		logger.Errorf("Failed to handle remote update: %v", err)
		return &pb.BoolResponse{Response: false}, err
	}
	m.vxlanManager.Add(pod.Vni, &pod.NetNs)

	elapsed := time.Since(startTime)
	m.latencyHistograms.Observe("remoteUpdate", elapsed.Milliseconds())
	logger.Infof("Successfully handled remote update in %v", elapsed)
	return &pb.BoolResponse{Response: true}, nil
}

// ------------------------------------------------------------------------------------------------------
func (m *KubeDTN) RemGRPCWire(ctx context.Context, wireDef *pb.WireDef) (*pb.BoolResponse, error) {
	if err := grpcwire.DeleteWiresByPod(wireDef.KubeNs, wireDef.LocalPodName); err != nil {
		return &pb.BoolResponse{Response: false}, err
	}
	return &pb.BoolResponse{Response: true}, nil
}

// ------------------------------------------------------------------------------------------------------
func (m *KubeDTN) AddGRPCWireLocal(ctx context.Context, wireDef *pb.WireDef) (*pb.BoolResponse, error) {
	logger := logger.WithFields(log.Fields{
		"overlay": "gRPC",
	})
	locInf, err := net.InterfaceByName(wireDef.VethNameLocalHost)
	if err != nil {
		logger.Errorf("[ADD-WIRE:LOCAL-END]For pod %s failed to retrieve interface ID for interface %v. error:%v", wireDef.LocalPodName, wireDef.VethNameLocalHost, err)
		return &pb.BoolResponse{Response: false}, err
	}

	//Using google gopacket for packet receive. An alternative could be using socket. Not sure it it provides any advantage over gopacket.
	wrHandle, err := pcap.OpenLive(wireDef.VethNameLocalHost, 65365, true, pcap.BlockForever)
	if err != nil {
		logger.Fatalf("[ADD-WIRE:LOCAL-END]Could not open interface for send/recv packets for containers. error:%v", err)
		return &pb.BoolResponse{Response: false}, err
	}

	aWire := grpcwire.GRPCWire{
		UID: int(wireDef.LinkUid),

		LocalNodeIfaceID:   int64(locInf.Index),
		LocalNodeIfaceName: wireDef.VethNameLocalHost,
		LocalPodIP:         wireDef.LocalPodIp,
		LocalPodIfaceName:  wireDef.IntfNameInPod,
		LocalPodName:       wireDef.LocalPodName,
		LocalPodNetNS:      wireDef.LocalPodNetNs,

		PeerIfaceID: wireDef.PeerIntfId,
		PeerNodeIP:  wireDef.PeerIp,

		Originator:   grpcwire.HOST_CREATED_WIRE,
		OriginatorIP: "unknown", /*+++todo retrieve host ip and set it here. Needed only for debugging */

		StopC:     make(chan struct{}),
		Namespace: wireDef.KubeNs,
	}

	grpcwire.AddWire(&aWire, wrHandle)

	logger.Infof("[ADD-WIRE:LOCAL-END]For pod %s@%s starting the local packet receive thread", wireDef.LocalPodName, wireDef.IntfNameInPod)
	// TODO: handle error here
	go grpcwire.RecvFrmLocalPodThread(&aWire)

	return &pb.BoolResponse{Response: true}, nil
}

// ------------------------------------------------------------------------------------------------------
func (m *KubeDTN) SendToOnce(ctx context.Context, pkt *pb.Packet) (*pb.BoolResponse, error) {
	logger := logger.WithFields(log.Fields{
		"overlay": "gRPC",
	})
	wrHandle, err := grpcwire.GetHostIntfHndl(pkt.RemotIntfId)
	if err != nil {
		logger.Errorf("SendToOnce (wire id - %v): Could not find local handle. err:%v", pkt.RemotIntfId, err)
		return &pb.BoolResponse{Response: false}, err
	}

	// In case any per packet log need to be generated.
	// pktType := grpcwire.DecodePkt(pkt.Frame)
	// logger.Printf("Daemon(SendToOnce): Received [pkt: %s, bytes: %d, for local interface id: %d]. Sending it to local container", pktType, len(pkt.Frame), pkt.RemotIntfId)
	// logger.Printf("Daemon(SendToOnce): Received [bytes: %d, for local interface id: %d]. Sending it to local container", len(pkt.Frame), pkt.RemotIntfId)

	err = wrHandle.WritePacketData(pkt.Frame)
	if err != nil {
		logger.Errorf("SendToOnce (wire id - %v): Could not write packet(%d bytes) to local interface. err:%v", pkt.RemotIntfId, len(pkt.Frame), err)
		return &pb.BoolResponse{Response: false}, err
	}

	return &pb.BoolResponse{Response: true}, nil
}

// ---------------------------------------------------------------------------------------------------------------
func (m *KubeDTN) AddGRPCWireRemote(ctx context.Context, wireDef *pb.WireDef) (*pb.WireCreateResponse, error) {
	stopC := make(chan struct{})
	wire, err := grpcwire.CreateGRPCWireRemoteTriggered(wireDef, stopC)

	if err == nil {
		logger.Infof("[ADD-WIRE:REMOTE-END]For pod %s@%s starting the local packet receive thread", wireDef.LocalPodName, wireDef.IntfNameInPod)
		go grpcwire.RecvFrmLocalPodThread(wire)

		return &pb.WireCreateResponse{Response: true, PeerIntfId: wire.LocalNodeIfaceID}, nil
	}
	logger.Errorf("[ADD-WIRE:REMOTE-END] err: %v", err)
	return &pb.WireCreateResponse{Response: false, PeerIntfId: wireDef.PeerIntfId}, err
}

// ---------------------------------------------------------------------------------------------------------------
// GRPCWireExists will return the wire if it exists.
func (m *KubeDTN) GRPCWireExists(ctx context.Context, wireDef *pb.WireDef) (*pb.WireCreateResponse, error) {
	wire, ok := grpcwire.GetWireByUID(wireDef.LocalPodNetNs, int(wireDef.LinkUid))
	if !ok || wire == nil {
		return &pb.WireCreateResponse{Response: false, PeerIntfId: wireDef.PeerIntfId}, nil
	}
	return &pb.WireCreateResponse{Response: ok, PeerIntfId: wire.PeerIfaceID}, nil
}

// ---------------------------------------------------------------------------------------------------------------
// Given the pod name and the pod interface, GenerateNodeInterfaceName generates the corresponding interface name in the node.
// This pod interface and the node interface later become the two end of a veth-pair
func (m *KubeDTN) GenerateNodeInterfaceName(ctx context.Context, in *pb.GenerateNodeInterfaceNameRequest) (*pb.GenerateNodeInterfaceNameResponse, error) {
	locIfNm, err := grpcwire.GenNodeIfaceName(in.PodName, in.PodIntfName)
	if err != nil {
		return &pb.GenerateNodeInterfaceNameResponse{Ok: false, NodeIntfName: ""}, err
	}
	return &pb.GenerateNodeInterfaceNameResponse{Ok: true, NodeIntfName: locIfNm}, nil
}

func (m *KubeDTN) addLink(ctx context.Context, localPod *pb.Pod, link *pb.Link) error {
	logger := common.GetLogger(ctx).WithFields(log.Fields{
		"link": link.Uid,
	})
	ctx = common.WithLogger(ctx, logger)
	ctx = common.WithCtxValue(ctx, common.TCPIP_BYPASS, m.config.TCPIPBypass)

	logger.Infof("Adding link: %v", link)
	startTime := time.Now()

	// Build koko's veth struct for local intf
	myVeth, err := common.MakeVeth(ctx, localPod.NetNs, link.LocalIntf, link.LocalIp, link.LocalMac)
	if err != nil {
		return err
	}

	// First option is macvlan interface
	if link.PeerPod == common.Localhost {
		logger.Infof("Peer link is MacVlan")
		macVlan := koko.MacVLan{
			ParentIF: link.PeerIntf,
			Mode:     common.MacvlanMode,
		}
		if err = koko.MakeMacVLan(*myVeth, macVlan); err != nil {
			logger.Infof("Failed to add macvlan interface")
			return err
		}
		logger.Infof("macvlan interfacee %s@%s has been added", link.LocalIntf, link.PeerIntf)
		return nil
	}

	// Physical-virtual link
	if strings.HasPrefix(link.PeerPod, "physical/") {
		srcIp := strings.TrimPrefix(link.PeerPod, "physical/")
		logger.Infof("Peer pod is physical host %s", srcIp)
		// Update local pod on behalf of the physical host
		// localPod is K8s pod, peerPod is physical host
		payload := &pb.RemotePod{
			NetNs:      localPod.NetNs,
			IntfName:   link.LocalIntf,
			IntfIp:     link.LocalIp,
			PeerVtep:   srcIp,
			Vni:        common.GetVniFromUid(link.Uid),
			KubeNs:     localPod.KubeNs,
			Properties: link.Properties,
			Name:       link.PeerPod,
		}
		response, err := m.Update(ctx, payload)
		if !response.Response || err != nil {
			return err
		}
		logger.Info("Successfully established physical-virtual link")
		return nil
	}

	// Virtual-virtual link
	// Initialising peer pod's metadata
	// We need original topology object here so avoid another query
	// to the API server in IsSkipped
	peerTopology, err := m.getPod(ctx, link.PeerPod, localPod.KubeNs)
	if err != nil {
		logger.Errorf("Failed to retrieve peer pod %s/%s topology", localPod.KubeNs, link.PeerPod)
		return err
	}
	peerPod, err := m.ToProtoPod(ctx, peerTopology)
	if err != nil {
		logger.Errorf("Failed to convert peer topology %s/%s to proto pod", localPod.KubeNs, link.PeerPod)
		return err
	}

	isAlive := peerPod.SrcIp != "" && peerPod.NetNs != ""
	logger.Infof("Is peer pod %s alive?: %t", peerPod.Name, isAlive)

	if !isAlive {
		// This means that our peer pod hasn't come up yet
		// Since there's no way of telling if our peer is going to be on this host or another,
		// the only option is to do nothing, assuming that the peer POD will do all the plumbing when it comes up
		logger.Infof("Peer pod %s isn't alive yet, continuing", peerPod.Name)
		return nil
	}

	// This means we're coming up AFTER our peer so things are pretty easy
	logger.Infof("Peer pod %s is alive", peerPod.Name)
	if peerPod.SrcIp == localPod.SrcIp { // This means we're on the same host
		logger.Infof("%s and %s are on the same host", localPod.Name, peerPod.Name)
		// Creating koko's Veth struct for peer intf
		peerVeth, err := common.MakeVeth(ctx, peerPod.NetNs, link.PeerIntf, link.PeerIp, link.PeerMac)
		if err != nil {
			logger.Errorf("Failed to build koko Veth struct")
			return err
		}

		mutex := m.linkMutexes.Get(link.Uid)
		mutex.Lock()
		defer mutex.Unlock()

		err = common.SetupVeth(ctx, myVeth, peerVeth, link, localPod)
		if err != nil {
			logger.Errorf("Error when creating a new VEth pair with koko: %s", err)
			logger.Infof("SELF VETH STRUCT: %+v", spew.Sdump(myVeth))
			logger.Infof("PEER VETH STRUCT: %+v", spew.Sdump(peerVeth))
			return err
		}
	} else { // This means we're on different hosts
		logger.Infof("%s@%s and %s@%s are on different hosts", localPod.Name, localPod.SrcIp, peerPod.Name, peerPod.SrcIp)

		vxlanSpec := &vxlan.VxlanSpec{
			NetNs:    localPod.NetNs,
			IntfName: link.LocalIntf,
			IntfIp:   link.LocalIp,
			PeerVtep: peerPod.SrcIp,
			Vni:      common.GetVniFromUid(link.Uid),
			SrcIp:    m.nodeIP,
			SrcIntf:  m.vxlanIntf,
		}

		mutex := m.linkMutexes.Get(link.Uid)
		mutex.Lock()

		if err = vxlan.SetupVxLan(ctx, vxlanSpec, link.Properties); err != nil {
			logger.Infof("Error when setting up VXLAN interface with koko: %s", err)
			mutex.Unlock()
			return err
		}
		m.vxlanManager.Add(vxlanSpec.Vni, &vxlanSpec.NetNs)

		// Unlock in advance to avoid deadlock
		// If we and remote daemon are both in `addLink` and called `UpdateRemote` simultaneously,
		// we will both be waiting forever since `Update` cannot acquire the lock
		// while `addLink` is holding it.
		mutex.Unlock()
		// Now we need to make an API call to update the remote VTEP to point to us
		err = common.UpdateRemote(ctx, localPod, peerPod, link)
		if err != nil {
			return err
		}
		logger.Infof("Successfully updated remote daemon")
	}

	elapsed := time.Since(startTime)
	m.latencyHistograms.Observe("add", elapsed.Milliseconds())
	logger.Infof("Successfully added link in %v", elapsed)
	return nil
}

func (m *KubeDTN) delLink(ctx context.Context, localPod *pb.Pod, link *pb.Link) error {
	logger := common.GetLogger(ctx).WithFields(log.Fields{
		"link": link.Uid,
	})
	ctx = common.WithLogger(ctx, logger)
	logger.Infof("Deleting link: %v", link)
	startTime := time.Now()

	// Creating koko's Veth struct for local intf
	myVeth, err := common.MakeVeth(ctx, localPod.NetNs, link.LocalIntf, link.LocalIp, link.LocalMac)
	if err != nil {
		logger.Infof("Failed to construct koko Veth struct")
		return err
	}

	// API call to koko to remove local Veth link
	if err = myVeth.RemoveVethLink(); err != nil {
		// instead of failing, just log the error and move on
		logger.Infof("Failed to remove veth link: %s", err)
	}

	vni := common.GetVniFromUid(link.Uid)
	netns := m.vxlanManager.Get(vni)
	if netns != nil && *netns == localPod.NetNs {
		m.vxlanManager.Delete(vni)
	}

	elapsed := time.Since(startTime)
	m.latencyHistograms.Observe("del", elapsed.Milliseconds())
	logger.Infof("Successfully deleted link in %v", elapsed)
	return nil
}

// Setup a pod, adding all its VXLAN VTEPs and links
func (m *KubeDTN) SetupPod(ctx context.Context, pod *pb.SetupPodQuery) (*pb.BoolResponse, error) {
	logger := common.GetLogger(ctx).WithFields(log.Fields{
		"pod":    pod.Name,
		"ns":     pod.KubeNs,
		"action": "setup",
	})
	ctx = common.WithLogger(ctx, logger)
	logger.Infof("Setting up pod")

	localPod, err := m.Get(ctx, &pb.PodQuery{
		Name:   string(pod.Name),
		KubeNs: string(pod.KubeNs),
	})
	if err != nil {
		logger.Infof("Pod is not in topology returning. err: %v", err)
		// Pod is not be in topology, the CNI plugin should delegate the action to the next plugin
		return &pb.BoolResponse{Response: true}, nil
	}

	// Marking pod as "alive" by setting its srcIP and NetNS
	localPod.NetNs = pod.NetNs
	localPod.SrcIp = m.nodeIP
	logger.Infof("Setting pod alive status")
	ok, err := m.SetAlive(ctx, localPod)
	if err != nil || !ok.Response {
		logger.Infof("Failed to set pod alive status: %v", err)
		return &pb.BoolResponse{Response: false}, err
	}

	response, err := m.AddLinks(ctx, &pb.LinksBatchQuery{
		LocalPod: localPod,
		Links:    localPod.Links,
	})
	if err != nil || !response.Response {
		logger.Infof("Failed to setup links: %v", err)
		return &pb.BoolResponse{Response: false}, err
	}

	logger.Infof("Successfully set up pod")
	return &pb.BoolResponse{Response: true}, nil
}

// Destroy a pod, removing all its GRPC wires and links, the reverse process of SetupPod
func (m *KubeDTN) DestroyPod(ctx context.Context, pod *pb.PodQuery) (*pb.BoolResponse, error) {
	logger := common.GetLogger(ctx).WithFields(log.Fields{
		"pod":    pod.Name,
		"ns":     pod.KubeNs,
		"action": "destroy",
	})
	ctx = common.WithLogger(ctx, logger)
	logger.Infof("Destroying pod")

	m.topologyManager.Delete(pod.Name, pod.KubeNs)
	// Close the grpc tunnel for this pod netns (if any)
	wireDef := pb.WireDef{
		KubeNs:       string(pod.KubeNs),
		LocalPodName: string(pod.Name),
	}

	removResp, err := m.RemGRPCWire(ctx, &wireDef)
	if err != nil || !removResp.Response {
		return &pb.BoolResponse{Response: false}, fmt.Errorf("could not remove grpc wire: %v", err)
	}

	localPod, err := m.Get(ctx, &pb.PodQuery{
		Name:   string(pod.Name),
		KubeNs: string(pod.KubeNs),
	})
	if err != nil {
		logger.Infof("Pod is not in topology returning. err: %v", err)
		// Pod is not be in topology, the CNI plugin should delegate the action to the next plugin
		// This is a special combination of return values
		return &pb.BoolResponse{Response: false}, nil
	}

	logger.Infof("Topology data still exists in CRs, cleaning up its status")
	// By setting srcIP and NetNS to "" we're marking this POD as dead
	localPod.NetNs = ""
	localPod.SrcIp = ""
	_, err = m.SetAlive(ctx, localPod)
	if err != nil {
		return &pb.BoolResponse{Response: false}, fmt.Errorf("could not set alive status: %v", err)
	}

	response, err := m.DelLinks(ctx, &pb.LinksBatchQuery{
		LocalPod: localPod,
		Links:    localPod.Links,
	})
	if err != nil || !response.Response {
		logger.Infof("Failed to destroy links: %v", err)
		return &pb.BoolResponse{Response: false}, err
	}

	logger.Infof("Successfully destroyed pod")
	return &pb.BoolResponse{Response: true}, nil
}

func (m *KubeDTN) AddLinks(ctx context.Context, query *pb.LinksBatchQuery) (*pb.BoolResponse, error) {
	localPod := query.LocalPod
	logger := common.GetLogger(ctx).WithFields(log.Fields{
		"pod":    localPod.Name,
		"ns":     localPod.KubeNs,
		"action": "add",
	})
	ctx = common.WithLogger(ctx, logger)

	for _, link := range query.Links {
		err := m.addLink(ctx, localPod, link)
		if err != nil {
			logger.WithField("link", link.Uid).Errorf("Failed to add link: %v", err)
			return &pb.BoolResponse{Response: false}, err
		}
	}

	logger.Infof("Successfully added links")
	return &pb.BoolResponse{Response: true}, nil
}

func (m *KubeDTN) DelLinks(ctx context.Context, query *pb.LinksBatchQuery) (*pb.BoolResponse, error) {
	localPod := query.LocalPod
	logger := logger.WithFields(log.Fields{
		"pod":    localPod.Name,
		"ns":     localPod.KubeNs,
		"action": "delete",
	})
	ctx = common.WithLogger(ctx, logger)

	for _, link := range query.Links {
		err := m.delLink(ctx, localPod, link)
		if err != nil {
			logger.WithField("link", link.Uid).Errorf("Failed to delete link: %v", err)
			return &pb.BoolResponse{Response: false}, err
		}
	}

	logger.Infof("Successfully deleted links")
	return &pb.BoolResponse{Response: true}, nil
}

func (m *KubeDTN) UpdateLinks(ctx context.Context, query *pb.LinksBatchQuery) (*pb.BoolResponse, error) {
	localPod := query.LocalPod
	logger := common.GetLogger(ctx).WithFields(log.Fields{
		"pod":    localPod.Name,
		"ns":     localPod.KubeNs,
		"action": "update",
	})
	ctx = common.WithLogger(ctx, logger)
	ctx = common.WithCtxValue(ctx, common.TCPIP_BYPASS, m.config.TCPIPBypass)

	for _, link := range query.Links {
		logger := logger.WithField("link", link.Uid)
		logger.Infof("Updating link")
		startTime := time.Now()

		myVeth, err := common.MakeVeth(ctx, localPod.NetNs, link.LocalIntf, link.LocalIp, link.LocalMac)
		if err != nil {
			return &pb.BoolResponse{Response: false}, err
		}
		qdiscs, err := common.MakeQdiscs(ctx, link.Properties)
		if err != nil {
			logger.Errorf("Failed to construct qdiscs: %s", err)
			return &pb.BoolResponse{Response: false}, err
		}
		err = common.SetVethQdiscs(ctx, myVeth, qdiscs)
		if err != nil {
			logger.Errorf("Failed to update qdiscs on self veth %s: %v", myVeth, err)
			return &pb.BoolResponse{Response: false}, err
		}

		elapsed := time.Since(startTime)
		m.latencyHistograms.Observe("update", elapsed.Milliseconds())
		logger.Infof("Successfully updated link in %v", elapsed)
	}

	logger.Infof("Successfully updated links")
	return &pb.BoolResponse{Response: true}, nil
}
