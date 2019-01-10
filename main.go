package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/disk"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/mackerelio/go-osstat/network"
	"github.com/mackerelio/go-osstat/uptime"
	// "reflect"
)

const (
	SLEEP   uint64 = 40
	VERSION uint64 = 1
)

type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

func main() {
	for true {
		loc, _ := time.LoadLocation("UTC")
		start := time.Now().In(loc)
		fmt.Println("Executing script", start)

		getSystemData()

		now := time.Now().In(loc)
		sleep_time := 60 - now.Second()

		if start.Minute() == now.Minute() {
			time.Sleep(time.Duration(sleep_time) * time.Second)
		}
	}
}

func getSystemData() {
	jsonData := map[string]interface{}{
		"version":       VERSION,
		"disk_all":      nil,
		"disk_used":     nil,
		"disk_free":     nil,
		"uptime":        nil,
		"memory_total":  nil,
		"memory_used":   nil,
		"memory_cached": nil,
		"memory_free":   nil,
		"cpu_user":      nil,
		"cpu_system":    nil,
		"cpu_idle":      nil,
		"network_io_tx": nil,
		"network_io_rx": nil,
		"disk_reads":    nil,
		"disk_writes":   nil,
	}

	before, err := cpu.Get()
	if err != nil {
		return
	}

	network_before, err := network.Get()
	if err != nil {
		return
	}

	disks_before, err := disk.Get()
	if err != nil {
		return
	}

	time.Sleep(time.Duration(SLEEP) * time.Second)

	after, err := cpu.Get()
	if err != nil {
		return
	}
	total := float64(after.Total - before.Total)

	network_after, err := network.Get()
	if err != nil {
		return
	}

	var network_io_transmit uint64 = 0
	var network_io_receive uint64 = 0

	for _, net_after := range network_after {
		for _, net_before := range network_before {
			if net_after.Name == net_before.Name {
				network_io_transmit = network_io_transmit + (net_after.TxBytes - net_before.TxBytes)
				network_io_receive = network_io_receive + (net_after.RxBytes - net_before.RxBytes)
			}
		}
	}

	jsonData["network_io_tx"] = network_io_transmit / SLEEP
	jsonData["network_io_rx"] = network_io_receive / SLEEP

	disks_after, err := disk.Get()
	if err != nil {
		return
	}

	var disk_io_reads uint64 = 0
	var disk_io_writes uint64 = 0

	for _, disk_after := range disks_after {
		for _, disk_before := range disks_before {
			if disk_after.Name == disk_before.Name {
				disk_io_reads = disk_io_reads + (disk_after.ReadsCompleted - disk_before.ReadsCompleted)
				disk_io_writes = disk_io_writes + (disk_after.WritesCompleted - disk_before.WritesCompleted)
			}
		}
	}

	jsonData["disk_reads"] = float64(disk_io_reads) / float64(SLEEP)
	jsonData["disk_writes"] = float64(disk_io_writes) / float64(SLEEP)

	// Disk
	disk := DiskUsage("/")
	jsonData["disk_all"] = float64(disk.All)
	jsonData["disk_used"] = float64(disk.Used)
	jsonData["disk_free"] = float64(disk.Free)

	// Uptime
	uptime, err := uptime.Get()
	if err != nil {
		return
	}
	jsonData["uptime"] = uptime

	// Memory
	memory, err := memory.Get()
	if err != nil {
		return
	}
	jsonData["memory_total"] = memory.Total
	jsonData["memory_used"] = memory.Used
	jsonData["memory_cached"] = memory.Cached
	jsonData["memory_free"] = memory.Free

	// CPU
	jsonData["cpu_user"] = float64(after.User-before.User) / total * 100
	jsonData["cpu_system"] = float64(after.System-before.System) / total * 100
	jsonData["cpu_idle"] = float64(after.Idle-before.Idle) / total * 100

	sendData(jsonData)
}

func sendData(jsonData map[string]interface{}) {
	jsonValue, _ := json.Marshal(jsonData)
	_, err := http.Post("http://ws.devect.com/api/v1/server/data/8f90fb8334c368bee02018c7e7e5de59/", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}
}

func DiskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}
