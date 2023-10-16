package pod

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
	"ws/dtn-satellite-sdn/sdn/link"
	"ws/dtn-satellite-sdn/sdn/util"
)

func GeneratePodSummaryFile(nameMap map[int]string, edgeSet []link.LinkEdge, outputPath string, expectedNodeNum int) error {
	podList := util.PodList{}
	podList.Kind = "PodList"
	podList.APIVersion = "v1"

	// Construct pods
	for idx := 0; idx < len(nameMap); idx++ {
		pod := util.Pod{
			APIVersion: "v1",
			Kind:       "Pod",
			MetaData: util.MetaData{
				Name: nameMap[idx],
			},
			Spec: util.PodSpec{
				Containers: []util.Container{
					{
						Name:            "satellite",
						Image:           "electronicwaste/podserver:v6",
						ImagePullPolicy: "IfNotPresent",
						Ports: []util.ContainerPort{
							{
								ContainerPort: 8080,
							},
						},
						Command: []string {
							"/bin/sh",
							"-c",
						},
						Args: []string {
							fmt.Sprintf(
								"ifconfig eth0:sdneth0 %s netmask 255.255.255.255 up;" + 
								"iptables -t nat -A OUTPUT -d 10.233.0.0/16 -j MARK --set-mark %d;" +
								"iptables -t nat -A POSTROUTING -m mark --mark %d -d 10.233.0.0/16 -j SNAT --to-source %s;" +
								"/podserver", 
								util.GetGlobalIP(uint(idx)),
								idx + 5000,
								idx + 5000,
								util.GetGlobalIP(uint(idx)),
							),
						},
						SecurityContext: util.SecurityContext{
							Capabilities: util.Capabilities{
								Add: []util.Capability{
									"NET_ADMIN",
								},
							},
						},
					},
				},
			},
		}
		podList.Items = append(podList.Items, pod)
	}

	// Write to file
	conf, err := yaml.Marshal(podList)
	if err != nil {
		return fmt.Errorf("Error in parsing podList to yaml: %v\n", err)
	}
	ioutil.WriteFile(outputPath, conf, 0644)

	return nil
}

