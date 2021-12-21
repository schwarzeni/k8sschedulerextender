package main

import (
	"fmt"
	"systeminfoagent/collector"
	"systeminfoagent/model"
)

func main() {
	// cpuinfo()
	// memoryinfo()
	// diskinfo()
	// netinfo()
	metric := &model.NodeMetric{}
	c := collector.NewDefaultCollector()
	c.Collect(metric)
	fmt.Printf("%+v\n", *metric)
}
