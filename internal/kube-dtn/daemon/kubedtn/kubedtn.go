package kubedtn

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	glogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	topologyclientv1 "github.com/y-young/kube-dtn/api/clientset/v1beta1"
	v1 "github.com/y-young/kube-dtn/api/v1"
	"github.com/y-young/kube-dtn/common"
	"github.com/y-young/kube-dtn/daemon/metrics"
	"github.com/y-young/kube-dtn/daemon/vxlan"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"

	pb "github.com/y-young/kube-dtn/proto/v1"
)

type Config struct {
	Port        int
	GRPCOpts    []grpc.ServerOption
	TCPIPBypass bool
}

type KubeDTN struct {
	pb.UnimplementedLocalServer
	pb.UnimplementedRemoteServer
	pb.UnimplementedWireProtocolServer
	config            Config
	kClient           kubernetes.Interface
	tClient           topologyclientv1.Interface
	rCfg              *rest.Config
	s                 *grpc.Server
	lis               net.Listener
	topologyStore     cache.Store
	topologyManager   *metrics.TopologyManager
	vxlanManager      *vxlan.VxlanManager
	latencyHistograms *metrics.LatencyHistograms
	linkMutexes       *common.MutexMap
	// IP of the node on which the daemon is running.
	nodeIP string
	// VXLAN interface name.
	vxlanIntf string
}

var logger *log.Entry = nil

func InitLogger() {
	logger = log.WithFields(log.Fields{"daemon": "kubedtnd"})
}

func restConfig() (*rest.Config, error) {
	logger.Infof("Trying in-cluster configuration")
	rCfg, err := rest.InClusterConfig()
	if err != nil {
		kubecfg := filepath.Join(".kube", "config")
		if home := homedir.HomeDir(); home != "" {
			kubecfg = filepath.Join(home, kubecfg)
		}
		logger.Infof("Falling back to kubeconfig: %q", kubecfg)
		rCfg, err = clientcmd.BuildConfigFromFlags("", kubecfg)
		if err != nil {
			return nil, err
		}
	}
	return rCfg, nil
}

func New(cfg Config, topologyManager *metrics.TopologyManager, latencyHistograms *metrics.LatencyHistograms) (*KubeDTN, error) {
	rCfg, err := restConfig()
	if err != nil {
		return nil, err
	}
	kClient, err := kubernetes.NewForConfig(rCfg)
	if err != nil {
		return nil, err
	}
	tClient, err := topologyclientv1.NewForConfig(rCfg)
	if err != nil {
		return nil, err
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, err
	}

	nodeIP := os.Getenv("HOST_IP")

	ctx := context.Background()
	topologies, err := tClient.Topology("").List(ctx, metav1.ListOptions{})
	logger.Infof("Found %d local topologies", len(topologies.Items))
	if err != nil {
		return nil, fmt.Errorf("failed to list topologies: %v", err)
	}
	localTopologies := filterLocalTopologies(topologies, &nodeIP)

	err = topologyManager.Init(localTopologies)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize topology manager: %v", err)
	}

	vxlanManager := vxlan.NewVxlanManager()
	vxlanManager.Init(localTopologies)

	_, vxlanIntf, err := vxlan.GetVxlanSource(nodeIP)
	if err != nil {
		return nil, fmt.Errorf("failed to get vxlan source: %v", err)
	}
	logger.Infof("Node IP: %s, VXLAN interface: %s", nodeIP, vxlanIntf)

	store, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return tClient.Topology("").List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return tClient.Topology("").Watch(ctx, options)
			},
		},
		&v1.Topology{},
		0,
		cache.ResourceEventHandlerFuncs{},
	)

	go controller.Run(wait.NeverStop)

	m := &KubeDTN{
		config:            cfg,
		rCfg:              rCfg,
		kClient:           kClient,
		tClient:           tClient,
		lis:               lis,
		s:                 newServerWithLogging(cfg.GRPCOpts...),
		topologyStore:     store,
		topologyManager:   topologyManager,
		vxlanManager:      vxlanManager,
		latencyHistograms: latencyHistograms,
		linkMutexes:       &common.MutexMap{},
		nodeIP:            nodeIP,
		vxlanIntf:         vxlanIntf,
	}
	pb.RegisterLocalServer(m.s, m)
	pb.RegisterRemoteServer(m.s, m)
	pb.RegisterWireProtocolServer(m.s, m)
	reflection.Register(m.s)
	return m, nil
}

func (m *KubeDTN) Serve() error {
	logger.Infof("GRPC server has started on port: %d", m.config.Port)
	return m.s.Serve(m.lis)
}

func (m *KubeDTN) Stop() {
	m.s.Stop()
}

func newServerWithLogging(opts ...grpc.ServerOption) *grpc.Server {
	lEntry := log.NewEntry(log.StandardLogger())
	lOpts := []glogrus.Option{}
	glogrus.ReplaceGrpcLogger(lEntry)
	opts = append(opts,
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			glogrus.UnaryServerInterceptor(lEntry, lOpts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			glogrus.StreamServerInterceptor(lEntry, lOpts...),
		))
	return grpc.NewServer(opts...)
}

func filterLocalTopologies(topologies *v1.TopologyList, nodeIP *string) *v1.TopologyList {
	filtered := &v1.TopologyList{}
	for _, topology := range topologies.Items {
		topology := topology
		if topology.Status.SrcIP == *nodeIP {
			filtered.Items = append(filtered.Items, topology)
		}
	}
	return filtered
}
