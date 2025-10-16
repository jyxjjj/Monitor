package collector

import (
	"runtime"
	"time"

	"github.com/jyxjjj/Monitor/pkg/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// Collector collects system metrics
type Collector struct {
	agentID      string
	lastNetRx    uint64
	lastNetTx    uint64
	lastNetTime  time.Time
}

// NewCollector creates a new metrics collector
func NewCollector(agentID string) *Collector {
	return &Collector{
		agentID:     agentID,
		lastNetTime: time.Now(),
	}
}

// Collect gathers current system metrics
func (c *Collector) Collect() (*models.Metrics, error) {
	metrics := &models.Metrics{
		AgentID:   c.agentID,
		Timestamp: time.Now(),
	}

	// CPU usage
	cpuPercents, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercents) > 0 {
		metrics.CPUPercent = cpuPercents[0]
	}

	// Memory usage
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metrics.MemoryUsed = memInfo.Used
		metrics.MemoryTotal = memInfo.Total
	}

	// Disk usage
	diskInfo, err := disk.Usage("/")
	if err == nil {
		metrics.DiskUsed = diskInfo.Used
		metrics.DiskTotal = diskInfo.Total
	}

	// Network usage
	netStats, err := net.IOCounters(false)
	if err == nil && len(netStats) > 0 {
		currentRx := netStats[0].BytesRecv
		currentTx := netStats[0].BytesSent
		
		// Calculate rate if we have previous data
		if c.lastNetRx > 0 {
			metrics.NetworkRx = currentRx - c.lastNetRx
			metrics.NetworkTx = currentTx - c.lastNetTx
		}
		
		c.lastNetRx = currentRx
		c.lastNetTx = currentTx
	}

	// Load average (not available on Windows)
	if runtime.GOOS != "windows" {
		loadInfo, err := load.Avg()
		if err == nil {
			metrics.LoadAvg1 = loadInfo.Load1
			metrics.LoadAvg5 = loadInfo.Load5
			metrics.LoadAvg15 = loadInfo.Load15
		}
	}

	return metrics, nil
}
