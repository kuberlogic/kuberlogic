package common

type Duration struct {
	MasterRead  int64
	MasterWrite int64
	ReplicaRead int64
}

type Delay struct {
	MasterRead  int64
	MasterWrite int64
	ReplicaRead int64
}
