package main

import (
	"fmt"
	"os"
	"systeminfoagent/diskusage"
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/disk"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/mackerelio/go-osstat/network"
)

func cpuinfo() {
	before, err := cpu.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	time.Sleep(time.Duration(1) * time.Second)
	after, err := cpu.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	total := float64(after.Total - before.Total)
	fmt.Printf("cpu user: %f %%\n", float64(after.User-before.User)/total*100)
	fmt.Printf("cpu system: %f %%\n", float64(after.System-before.System)/total*100)
	fmt.Printf("cpu idle: %f %%\n", float64(after.Idle-before.Idle)/total*100)
}

func memoryinfo() {
	memory, err := memory.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	fmt.Printf("memory total: %d bytes\n", memory.Total)
	fmt.Printf("memory used: %d bytes\n", memory.Used)
	fmt.Printf("memory cached: %d bytes\n", memory.Cached)
	fmt.Printf("memory free: %d bytes\n", memory.Free)

	// used / total
}

// diskinfo
// 目前可用的空间
// 读写的次数(1s内)
func diskinfo() {
	disks, err := disk.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "err : %s\n", err)
		return
	}
	var beforeR, beforeW, flag uint64
	for _, di := range disks {
		if di.Name == "sda" {
			beforeR, beforeW, flag = di.ReadsCompleted, di.WritesCompleted, 1
		}
	}
	if flag == 0 {
		fmt.Fprintf(os.Stderr, "sda not found\n")
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
	var KB = uint64(1024)

	fmt.Printf("read times in 1s: %d\n", afterR-beforeR)
	fmt.Printf("write times in 1s: %d\n", afterW-beforeW)
	fmt.Println("Free:", usage.Free()/(KB*KB))
	fmt.Println("Available:", usage.Available()/(KB*KB))
	fmt.Println("Size:", usage.Size()/(KB*KB))
	fmt.Println("Used:", usage.Used()/(KB*KB))
	fmt.Println("Usage:", usage.Usage()*100, "%")
}

func netinfo() {
	var (
		mainEthStatBefore network.Stats
		mainEthStatAfter  network.Stats
		mainEthName       = "enp0s5"
	)
	before, err := network.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "err : %s\n", err)
		return
	}
	for _, network := range before {
		if network.Name == mainEthName {
			mainEthStatBefore = network
		}
	}
	time.Sleep(time.Second)
	after, err := network.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "err : %s\n", err)
		return
	}
	for _, network := range after {
		if network.Name == mainEthName {
			mainEthStatAfter = network
		}
	}
	mainEthStatAfter.RxBytes = mainEthStatAfter.RxBytes - mainEthStatBefore.RxBytes
	mainEthStatAfter.TxBytes = mainEthStatAfter.TxBytes - mainEthStatBefore.TxBytes
	fmt.Printf("%s %d %d\n", mainEthName, mainEthStatAfter.RxBytes, mainEthStatAfter.TxBytes)
}
