package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/mackerelio/go-osstat/uptime"
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
	}

	before, err := cpu.Get()
	if err != nil {
		return
	}

	time.Sleep(time.Duration(40) * time.Second)

	after, err := cpu.Get()
	if err != nil {
		return
	}
	total := float64(after.Total - before.Total)

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
	_, err := http.Post("URL_API_HERE", "application/json", bytes.NewBuffer(jsonValue))
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
