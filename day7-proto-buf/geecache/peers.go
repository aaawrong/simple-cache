package geecache

import pb "day7/geecache/geecachepb"

type PeerPicker interface {
	// PickPeer() 方法用于根据传入的 key 选择相应节点 PeerGetter。
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	// Get() 方法用于从对应 group 查找缓存值
	//Get(group string, key string) ([]byte, error)
	Get(in *pb.Request, out *pb.Response) error
}
