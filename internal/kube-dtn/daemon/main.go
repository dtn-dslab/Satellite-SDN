package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/y-young/kube-dtn/bpf"
	"github.com/y-young/kube-dtn/common"
	"github.com/y-young/kube-dtn/daemon/cni"
	"github.com/y-young/kube-dtn/daemon/grpcwire"
	"github.com/y-young/kube-dtn/daemon/kubedtn"
	"github.com/y-young/kube-dtn/daemon/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := cni.Init(); err != nil {
		log.Errorf("Failed to initialise CNI plugin: %v", err)
		os.Exit(1)
	}
	defer cni.Cleanup()

	isDebug := flag.Bool("d", false, "enable degugging")
	grpcPort, err := strconv.Atoi(os.Getenv("GRPC_PORT"))
	if err != nil || grpcPort == 0 {
		grpcPort, err = strconv.Atoi(common.DefaultPort)
		if err != nil {
			panic("Failed to parse default port")
		}
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if err != nil || httpAddr == "" {
		httpAddr = common.HttpAddr
	}

	flag.Parse()
	log.SetLevel(log.InfoLevel)
	if *isDebug {
		log.SetLevel(log.DebugLevel)
		log.Debug("Verbose logging enabled")
	}

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		ForceColors:     true,
		FullTimestamp:   true,
	})

	kubedtn.InitLogger()
	grpcwire.InitLogger()

	topologyManager := metrics.NewTopologyManager()
	reg := metrics.NewRegistry(topologyManager)
	latencyHistograms := metrics.NewLatencyHistograms()
	latencyHistograms.Register(reg)

	go func() {
		log.Infof("HTTP server listening on port %s", httpAddr)
		http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		http.ListenAndServe(httpAddr, nil)
	}()

	tcpIpBypass := os.Getenv(common.TCPIP_BYPASS) == "1"
	if tcpIpBypass {
		log.Info("TCP/IP bypass enabled")
		var prog bpf.BypassProgram

		err = bpf.CheckOrMountBPFFSDefault()
		if err != nil {
			log.Errorf("BPF filesystem mounting on /sys/fs/bpf failed: %s", err)
			return
		}

		if err := bpf.SetLimit(); err != nil {
			log.Errorf("Setting limit failed: %s", err)
			return
		}

		prog, err = bpf.LoadProgram(prog)
		if err != nil {
			log.Errorf("Loading program failed: %s", err)
			return
		}
		defer bpf.CloseProgram(prog)
	}

	m, err := kubedtn.New(kubedtn.Config{
		Port:        grpcPort,
		TCPIPBypass: tcpIpBypass,
	}, topologyManager, latencyHistograms)
	if err != nil {
		log.Errorf("Failed to create kubedtn: %v", err)
		os.Exit(1)
	}
	log.Info("Starting kubedtn daemon...with grpc support")
	log.Infof("GRPC listening on port %d", grpcPort)

	if err := m.Serve(); err != nil {
		log.Errorf("Daemon exited badly: %v", err)
		os.Exit(1)
	}
}
