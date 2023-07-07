package GeeCache

// PeerGetter 用于从对应group查找缓存值，会应用于HTTP客户端
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

// PeerPicker 用于根据传入的key选择相应节点PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}
