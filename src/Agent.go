//go:build Agent

package main

import (
	"flag"
	"net/url"
	"time"
)

var URL string

func main() {
	flag.StringVar(&URL, "h", "status.desmg.com", "Host")
	flag.Parse()
	var scheme string
	if URL == "status.desmg.com:443" {
		scheme = "wss"
	} else {
		scheme = "ws"
	}
	go reconnectWebSocket(url.URL{Scheme: scheme, Host: URL, Path: "/agent"})
	defer closeWebSocket()
	DeviceID := GetDeviceID()
	for {
		data := run()
		send(DeviceID, data)
		time.Sleep(time.Second)
	}
}

func run() MonitorData {
	iorw := GetIORW()
	data := MonitorData{
		CPUModel:   GetCPUModel(),
		CPUNum:     GetCPUNum(),
		CPUFreq:    GetCPUFreq(),
		CPUUsage:   GetCPUUsage(),
		MemSize:    GetMemSize(),
		MemUsed:    GetMemUsed(),
		NumProcess: GetNumProcess(),
		DiskName:   GetDiskName(),
		DiskUsage:  GetDiskUsage(),
		DiskSize:   GetDiskSize(),
		Uptime:     GetUptime(),
		IORead:     iorw[0],
		IOWrite:    iorw[1],
	}
	return data
}
