package common

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/y-young/kube-dtn/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

type MutexMap sync.Map

func (m *MutexMap) Get(key any) *sync.Mutex {
	mutex, _ := (*sync.Map)(m).LoadOrStore(key, &sync.Mutex{})
	return mutex.(*sync.Mutex)
}

// Generate VXLAN Vni from link UID
func GetVniFromUid(uid int64) int32 {
	return int32(VxlanBase + uid)
}

// Get link UID from VXLAN Vni
func GetUidFromVni(vni int32) int64 {
	return int64(vni - VxlanBase)
}

// Call remote daemon to set up link on their side
func UpdateRemote(ctx context.Context, localPod *pb.Pod, peerPod *pb.Pod, link *pb.Link) error {
	logger := GetLogger(ctx)

	payload := &pb.RemotePod{
		NetNs:      peerPod.NetNs,
		IntfName:   link.PeerIntf,
		IntfIp:     link.PeerIp,
		PeerVtep:   localPod.SrcIp,
		Vni:        GetVniFromUid(link.Uid),
		KubeNs:     localPod.KubeNs,
		Properties: link.Properties,
		Name:       link.PeerPod,
	}

	url := fmt.Sprintf("passthrough:///%s:%s", peerPod.SrcIp, DefaultPort)
	logger.Infof("Trying to do a remote update on %s", url)

	remote, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Infof("Failed to dial remote gRPC url %s", url)
		return err
	}
	remoteClient := pb.NewRemoteClient(remote)
	ok, err := remoteClient.Update(ctx, payload)
	if err != nil || !ok.Response {
		logger.Infof("Failed to do a remote update: %s", err)
		return err
	}
	return nil
}
