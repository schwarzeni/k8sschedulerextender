package collector

import (
	"fmt"
	"log"
	"sync"
	"systeminfoagent/diskusage"
	"systeminfoagent/model"
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/disk"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/mackerelio/go-osstat/network"
)

// Collector 用于 agent 在节点上收集系统的相关信息
type Collector interface {
	Collect(*model.NodeMetric) error
}

type DefaultCollector struct {
	collectors []Collector
}

func NewDefaultCollector() *DefaultCollector {
	return &DefaultCollector{
		collectors: []Collector{&CPUCollector{}, &MemoryCollector{}, &DiskCollector{}, &NetCollector{}},
	}
}

func (dc *DefaultCollector) Collect(metric *model.NodeMetric) error {
	var wg = sync.WaitGroup{}
	metric.Timestamp = time.Now()
	wg.Add(len(dc.collectors))
	for _, collector := range dc.collectors {
		go func(collector Collector) {
			if err := collector.Collect(metric); err != nil {
				log.Println(err)
			}
			wg.Done()
		}(collector)
	}
	wg.Wait()
	return nil
}

type CPUCollector struct{}

func (*CPUCollector) Collect(metric *model.NodeMetric) error {
	before, err := cpu.Get()
	if err != nil {
		return fmt.Errorf("get cpu info: %v", err)
	}
	time.Sleep(time.Duration(1) * time.Second)
	after, err := cpu.Get()
	if err != nil {
		return fmt.Errorf("get cpu info: %v", err)
	}
	metric.CPU = model.CPU{
		Valid:  true,
		User:   after.User - before.User,
		System: after.System - before.System,
		Idle:   after.Idle - before.Idle,
	}
	return nil
}

type MemoryCollector struct{}

func (*MemoryCollector) Collect(metric *model.NodeMetric) error {
	memory, err := memory.Get()
	if err != nil {
		return fmt.Errorf("get memory info: %v", err)
	}
	metric.Memory = model.Memory{
		Valid:  true,
		Total:  memory.Total,
		Used:   memory.Used,
		Cached: memory.Cached,
		Free:   memory.Free,
	}
	return nil
}

type DiskCollector struct{}

func (*DiskCollector) Collect(metric *model.NodeMetric) error {
	disks, err := disk.Get()
	if err != nil {
		return fmt.Errorf("get diskinfo: %v", err)
	}
	var beforeR, beforeW, flag uint64
	for _, di := range disks {
		if di.Name == "sda" { // TODO: set as arg
			beforeR, beforeW, flag = di.ReadsCompleted, di.WritesCompleted, 1
		}
	}
	if flag == 0 {
		return fmt.Errorf("get diskinfo: sda not found")
	}
	time.Sleep(time.Duration(1) * time.Second)
	var afterR, afterW uint64
	disks, _ = disk.Get()
	for _, di := range disks {
		if di.Name == "sda" {
			afterR, afterW = di.ReadsCompleted, di.WritesCompleted
		}
	}
	usage := diskusage.NewDiskUsage("/")
	metric.Disk = model.Disk{
		Valid:      true,
		Size:       usage.Size(),
		Used:       usage.Used(),
		Free:       usage.Free(),
		WriteTimes: afterW - beforeW,
		ReadTimes:  afterR - beforeR,
	}
	return nil
}

type NetCollector struct{}

func (*NetCollector) Collect(metric *model.NodeMetric) error {
	var (
		mainEthStatBefore network.Stats
		mainEthStatAfter  network.Stats
		mainEthName       = "enp0s5" // TODO: set as arg
	)
	before, err := network.Get()
	if err != nil {
		return fmt.Errorf("get net: %v", err)
	}
	for _, network := range before {
		if network.Name == mainEthName {
			mainEthStatBefore = network
		}
	}
	time.Sleep(time.Second)
	after, err := network.Get()
	if err != nil {
		return fmt.Errorf("get net: %v", err)
	}
	for _, network := range after {
		if network.Name == mainEthName {
			mainEthStatAfter = network
		}
	}
	metric.Network = model.Network{
		Valid:   true,
		RxBytes: mainEthStatAfter.RxBytes - mainEthStatBefore.RxBytes,
		TxBytes: mainEthStatAfter.TxBytes - mainEthStatBefore.TxBytes,
	}
	return nil
}
