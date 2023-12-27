package metrics

import (
	"errors"
	"fmt"
	"math/rand"
)

const (
	StarMetrics = "StarStatus"
	LinkMetrics = "LinkStatus"
)

func GetFakeStarMetrics(indexUUIDMap map[int]string) []string {
	containerTypes := []string{"networkName", "Name", "CPU", "memory", "mempct"}
	result := []string{}
	for _, uuid := range indexUUIDMap {
		cpu, memory := randFloats(0.10, 1.00), randFloats(20.0, 40.0)
		mempct :=  100.0 * memory / (128 * 1024)
		containerData := []string{
			"Network1", uuid,
			fmt.Sprintf("%.2f%%", cpu),
			fmt.Sprintf("%.2fMiB", memory),
			fmt.Sprintf("%.2f%%", mempct),
		}
		if str, err := GeneratePrometheusData(StarMetrics, containerTypes, containerData); err == nil {
			result = append(result, str)
		}
	}
	return result
}

func GetFakeLinkMetrics(topoArr [][]string) []string {
	linkTypes := []string{"networkName", "pod", "star1", "star2", "Rbandwidth", "Tbandwidth"}
	result := []string{}
	for _, linkPair := range topoArr {
		from, to := linkPair[0], linkPair[1]
		rband, tband := randFloats(10.0, 40.0), randFloats(10.0, 40.0)
		linkData := []string{
			"Network1", from, from, to,
			fmt.Sprintf("%.2fMbps", rband),
			fmt.Sprintf("%.2fMbps", tband),
		}
		if str, err := GeneratePrometheusData(LinkMetrics, linkTypes, linkData); err == nil {
			result = append(result, str)
		}
		linkData = []string{
			"Network1", to, to, from,
			fmt.Sprintf("%.2fMbps", tband),
			fmt.Sprintf("%.2fMbps", rband),
		}
		if str, err := GeneratePrometheusData(LinkMetrics, linkTypes, linkData); err == nil {
			result = append(result, str)
		}
	}
	return result
}

func randFloats(min, max float64) float64 {
	return min + (max - min) * rand.Float64()
}

func GeneratePrometheusData(name string, types []string, datas []string) (string, error) {
	var result string
	if len(types) != len(datas) {
		return result, errors.New("lens of types is difference from lens of datas for prometheus type")
	}
	result += name + "{"
	for index := range types {
		result += fmt.Sprintf("%s=\"%s\",", types[index], datas[index])
	}
	result = result[:len(result)-1] + "} 1"
	return result, nil
}

