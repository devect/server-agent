package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/disk"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/mackerelio/go-osstat/network"
	"github.com/mackerelio/go-osstat/uptime"
	"github.com/maguayo/goInfo"
	"github.com/shirou/gopsutil/load"
	// "reflect"
)

const (
	SLEEP   uint64 = 40
	VERSION uint64 = 1
)

var GoOS string
var Kernel string
var Core string
var Platform string
var OS string
var Hostname string
var CPUs int

type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

func main() {
	args := os.Args[1:]
	if len(os.Args) > 1 {
		if args[0] == "credentials" {
			err := ioutil.WriteFile("/etc/devect/devect-auth.txt", []byte(args[1]), 0777)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}

	server_id, err := getServerId()
	server_id = strings.Replace(server_id, "\n", "", -1)

	if err != nil || IsValidUUID(server_id) != true {
		fmt.Println("You don't have credentials. \nRun: devect credentials 'YOUR_SERVER_ID_HERE'")
	} else {
		gi := goInfo.GetInfo()
		Kernel = gi.Kernel
		Core = gi.Core
		OS = gi.OS
		Hostname = gi.Hostname
		CPUs = gi.CPUs
		runLoop(server_id)
	}

}

func getServerId() (string, error) {
	b, err := ioutil.ReadFile("/etc/devect/devect-auth.txt")
	str := string(b)
	return str, err
}

func runLoop(server_id string) {
	for true {
		loc, _ := time.LoadLocation("UTC")
		start := time.Now().In(loc)
		fmt.Println("Executing script", start)

		getSystemData(server_id)

		now := time.Now().In(loc)
		sleep_time := 60 - now.Second()

		if start.Minute() == now.Minute() {
			time.Sleep(time.Duration(sleep_time) * time.Second)
		}
	}
}

func getSystemData(server_id string) {
	jsonData := map[string]interface{}{
		"agent_version": VERSION,
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
		"load_average":  nil,
		"hostname":      Hostname,
		"os":            OS,
		"linux_kernel":  Kernel,
		"cpu_name":      Core,
		"cpu_cores":     CPUs,
		"ip":            "127.0.0.1",
	}

	before, _ := cpu.Get()
	network_before, _ := network.Get()
	disks_before, _ := disk.Get()

	time.Sleep(time.Duration(SLEEP) * time.Second)

	after, _ := cpu.Get()
	total := float64(after.Total - before.Total)
	network_after, _ := network.Get()
	disks_after, _ := disk.Get()
	disk := DiskUsage("/")
	uptime, _ := uptime.Get()
	memory, _ := memory.Get()

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
	jsonData["disk_all"] = float64(disk.All)
	jsonData["disk_used"] = float64(disk.Used)
	jsonData["disk_free"] = float64(disk.Free)

	// Uptime
	jsonData["uptime"] = uptime

	// Memory
	jsonData["memory_total"] = memory.Total
	jsonData["memory_used"] = memory.Used
	jsonData["memory_cached"] = memory.Cached
	jsonData["memory_free"] = memory.Free

	// CPU
	jsonData["cpu_user"] = float64(after.User-before.User) / total * 100
	jsonData["cpu_system"] = float64(after.System-before.System) / total * 100
	jsonData["cpu_idle"] = float64(after.Idle-before.Idle) / total * 100

	// Load Average
	load, _ := load.Avg()
	jsonData["load_average"] = load.Load1

	sendData(server_id, jsonData)
}

func sendData(server_id string, jsonData map[string]interface{}) {
	fmt.Println(jsonData)
	jsonValue, _ := json.Marshal(jsonData)
	url := "http://api.devect.com/api/v1/server/" + server_id + "/"
	fmt.Println(url)
	_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
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

func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}
