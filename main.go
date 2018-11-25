package main

import (
	"os"
	"strconv"
	"time"

	"github.com/mackerelio/go-osstat/network"

	"github.com/mackerelio/go-osstat/disk"
)

func main() {
	f, _ := os.OpenFile("/usr/local/bin/drive.txt", os.O_APPEND|os.O_WRONLY, 0644)
	f2, _ := os.OpenFile("/usr/local/bin/network.txt", os.O_APPEND|os.O_WRONLY, 0644)

	for true {
		ds := getDisk()
		nt := getNet()
		currentTime := time.Now().String()
		under := ""
		for i := -3; i < len(currentTime); i++ {
			under += "_"
		}
		under += "\n"
		//println(under + "| " + currentTime + " |\n" + under + ds + "------------------------------------------------------------\n" + nt)

		f.WriteString(under + "| " + currentTime + " |\n" + under + ds + "------------------------------------------------------------\n")
		f2.WriteString(under + "| " + currentTime + " |\n" + under + nt + "------------------------------------------------------------\n")

		time.Sleep(30 * time.Second)

	}
}
func getDisk() string {
	txt := ""
	usage, _ := disk.Get()
	totalRead := 0
	totalWrite := 0
	for i := 0; i < len(usage); i++ {
		read := strconv.Itoa(int(usage[i].ReadsCompleted))
		write := strconv.Itoa(int(usage[i].WritesCompleted))
		txt += ("Drive : " + usage[i].Name + " | Reads : " + read + " | Writes: " + write + "\n")
		totalRead += int(usage[i].ReadsCompleted)
		totalWrite += int(usage[i].WritesCompleted)
	}
	txt += ("Total Read : " + strconv.Itoa(totalRead) + " | Total Write : " + strconv.Itoa(totalWrite) + "\n")
	return txt
}
func getNet() string {
	txt := ""
	usage, _ := (network.Get())
	totalDown := 0
	totalUp := 0
	for i := 0; i < len(usage); i++ {
		Down := strconv.Itoa(int(usage[i].RxBytes))
		Up := strconv.Itoa(int(usage[i].TxBytes))
		txt += ("Network Device : " + usage[i].Name + " | Download : " + Down + " | Upload: " + Up + "\n")
		totalDown += int(usage[i].RxBytes)
		totalUp += int(usage[i].TxBytes)
	}
	txt += ("Total Download : " + strconv.Itoa(totalDown) + " | Total Upload : " + strconv.Itoa(totalUp) + "\n")
	return txt
}
