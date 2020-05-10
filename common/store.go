package common

//存储类型，表示文件存储到哪里
type StoreType int

const (
	_ StoreType = iota
	// 节点本地
	StoreLocal
	// ceph集群
	StoreCeph
	// 阿里云OSS
	StoreOSS
	// 混合存储(ceph/oss)
	StoreMix
	// 所有存储类型都存储一份数据
	StoreAll
)
