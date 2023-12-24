package link

import (
	"testing"

	"ws/dtn-satellite-sdn/sdn/util"
)

func TestGenerateIP(t *testing.T) {
	ip := util.GetVxlanIP(1, 2)
	t.Logf("IP is %s\n", ip)
	if ip != "128.16.2.1/24" {
		t.Errorf("IP Dismatch!\n")
	}
}
