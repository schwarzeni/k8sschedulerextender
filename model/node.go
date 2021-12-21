package model

import "time"

// NodeInfoRecord 存储在 master 上的节点信息
type NodeInfoRecord struct {
	ID           string           `json:"id"`
	DownDuration time.Duration    `json:"down_duration"`
	Metrics      []NodeFullMetric `json:"metrics"`
}

type NodeFullMetric struct {
	RawMetric NodeMetric `json:"raw_metric"`
	NodeInfo  NodeInfo   `json:"node_info"`
	// 加入了平均数、方差等统计要素
	Statistics Statistics `json:"statistics"`
}

type Statistics struct {
	CPU     MetricStatistics `json:"cpu"`
	Memory  MetricStatistics `json:"memory"`
	Network MetricStatistics `json:"network"`
	Disk    MetricStatistics `json:"disk"`
}

type MetricStatistics struct {
	Mean     float64 `json:"mean"`
	Variance float64 `json:"variance"`
	N        int     `json:"n"` // 如何是 rawdata 是 invalid，则使用之前最新的不是 invalid 的数据
}

type NodeMetric struct {
	// 时间戳
	// 节点基本信息
	// 各种数据
	// 这里 Network 仅考虑单网卡的情况
	Timestamp time.Time `json:"timestamp"`
	NodeInfo  NodeInfo  `json:"node_info"`
	CPU       CPU       `json:"cpu"`
	Memory    Memory    `json:"memory"`
	Network   Network   `json:"network"`
	Disk      Disk      `json:"disk"`
}

type NodeInfo struct {
	ID string `json:"id"`
}

type CPU struct {
	Valid  bool   `json:"valid"`
	User   uint64 `json:"user"`
	System uint64 `json:"system"`
	Idle   uint64 `json:"idle"`
}

type Memory struct {
	Valid  bool   `json:"valid"`
	Total  uint64 `json:"total"`
	Used   uint64 `json:"used"`
	Cached uint64 `json:"cached"`
	Free   uint64 `json:"free"`
}

type Network struct {
	Valid   bool   `json:"valid"`
	RxBytes uint64 `json:"rx_bytes"`
	TxBytes uint64 `json:"tx_bytes"`
}

type Disk struct {
	Valid      bool   `json:"valid"`
	Size       uint64 `json:"size"`
	Used       uint64 `json:"used"`
	Free       uint64 `json:"free"`
	WriteTimes uint64 `json:"write_times"`
	ReadTimes  uint64 `json:"read_times"`
}
