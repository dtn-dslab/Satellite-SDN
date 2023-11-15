package metrics

import (
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	v1 "github.com/y-young/kube-dtn/api/v1"
)

type InterfaceStatisticsCollector struct {
	prometheus.Collector
	manager *TopologyManager
}

var (
	labels = []string{"interface", "pod", "namespace"}

	rxPacketsDesc = prometheus.NewDesc("interface_rx_packets",
		"Number of good packets received by the interface",
		labels,
		nil,
	)

	rxBytesDesc = prometheus.NewDesc("interface_rx_bytes",
		"Number of good received bytes, corresponding to rx_packets",
		labels,
		nil,
	)

	txPacketsDesc = prometheus.NewDesc("interface_tx_packets",
		"Number of packets successfully transmitted",
		labels,
		nil,
	)

	txBytesDesc = prometheus.NewDesc("interface_tx_bytes",
		"Number of good transmitted bytes, corresponding to tx_packets",
		labels,
		nil,
	)

	rxErrorsDesc = prometheus.NewDesc("interface_rx_errors",
		"Total number of bad packets received on this network device",
		labels,
		nil,
	)

	txErrorsDesc = prometheus.NewDesc("interface_tx_errors",
		"Total number of transmit problems",
		labels,
		nil,
	)

	rxDroppedDesc = prometheus.NewDesc("interface_rx_dropped",
		"Number of packets received but not processed, e.g. due to lack of resources or unsupported protocol",
		labels,
		nil,
	)

	txDroppedDesc = prometheus.NewDesc("interface_tx_dropped",
		"Number of packets dropped on their way to transmission, e.g. due to lack of resources",
		labels,
		nil,
	)
)

func (c *InterfaceStatisticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- rxPacketsDesc
	ch <- rxBytesDesc
	ch <- txPacketsDesc
	ch <- txBytesDesc
	ch <- rxErrorsDesc
	ch <- txErrorsDesc
	ch <- rxDroppedDesc
	ch <- txDroppedDesc
}

func (c *InterfaceStatisticsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, topology := range c.manager.List() {
		if topology.Status.NetNs == "" {
			continue
		}
		c.collectIpRouteMetrics(ch, topology)
	}
}

func (c *InterfaceStatisticsCollector) listLinks(nsName string) (links []netlink.Link, err error) {
	var vethNs ns.NetNS
	if nsName == "" {
		if vethNs, err = ns.GetCurrentNS(); err != nil {
			log.Errorf("Failed to get current network namespace: %v", err)
			return nil, err
		}
	} else {
		if vethNs, err = ns.GetNS(nsName); err != nil {
			log.Errorf("Failed to get network namespace %s: %v", nsName, err)
			return nil, err
		}
	}
	defer vethNs.Close()

	err = vethNs.Do(func(_ ns.NetNS) error {
		links, err = netlink.LinkList()
		return err
	})
	return links, err
}

func (c *InterfaceStatisticsCollector) collectIpRouteMetrics(ch chan<- prometheus.Metric, topology *v1.Topology) {
	links, err := c.listLinks(topology.Status.NetNs)
	if err != nil {
		log.WithFields(log.Fields{
			"topology": topology.Name,
			"ns":       topology.Namespace,
		}).Errorf("Failed to list links: %v", err)
		return
	}

	for _, link := range links {
		attrs := link.Attrs()
		stats := attrs.Statistics
		labels := []string{attrs.Name, topology.Name, topology.Namespace}
		ch <- prometheus.MustNewConstMetric(rxBytesDesc, prometheus.CounterValue, float64(stats.RxBytes), labels...)
		ch <- prometheus.MustNewConstMetric(rxPacketsDesc, prometheus.CounterValue, float64(stats.RxPackets), labels...)
		ch <- prometheus.MustNewConstMetric(txBytesDesc, prometheus.CounterValue, float64(stats.TxBytes), labels...)
		ch <- prometheus.MustNewConstMetric(txPacketsDesc, prometheus.CounterValue, float64(stats.TxPackets), labels...)
		ch <- prometheus.MustNewConstMetric(rxErrorsDesc, prometheus.CounterValue, float64(stats.RxErrors), labels...)
		ch <- prometheus.MustNewConstMetric(txErrorsDesc, prometheus.CounterValue, float64(stats.TxErrors), labels...)
		ch <- prometheus.MustNewConstMetric(rxDroppedDesc, prometheus.CounterValue, float64(stats.RxDropped), labels...)
		ch <- prometheus.MustNewConstMetric(txDroppedDesc, prometheus.CounterValue, float64(stats.TxDropped), labels...)
	}
}

func NewInterfaceStatisticsCollector(manager *TopologyManager) *InterfaceStatisticsCollector {
	return &InterfaceStatisticsCollector{
		manager: manager,
	}
}
