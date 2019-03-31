package monitoring

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/mackerelio/go-osstat/disk"
	"github.com/mackerelio/go-osstat/memory"
	log "github.com/sirupsen/logrus"
)

type DiskUsagePerNode struct {
	ID     uint64 `gorm:"primary_key"`
	Kind   string `gorm:"type:varchar(255) not null"`
	NodeID string `gorm:"type:varchar(255) not null"`

	DiskAll  uint64
	DiskUsed uint64
	DiskFree uint64

	CreatedAt time.Time
	UpdatedAt time.Time
}

type DiskIOPerNode struct {
	ID     uint64 `gorm:"primary_key"`
	Kind   string `gorm:"type:varchar(255) not null"`
	NodeID string `gorm:"type:varchar(255) not null"`

	DiskReadsCompleted  uint64
	DiskWritesCompleted uint64

	CreatedAt time.Time
}

type DiskMDPerNode struct {
	ID     uint64 `gorm:"primary_key"`
	Kind   string `gorm:"type:varchar(255) not null"`
	NodeID string `gorm:"type:varchar(255) not null"`

	ExitCode int
	MDADM    string `gorm:"type:text not null"`

	CreatedAt time.Time
}

type MemStatsPerNode struct {
	ID     uint64 `gorm:"primary_key"`
	NodeID string `gorm:"type:varchar(255) not null"`

	MemoryTotal      uint64
	MemoryUsed       uint64
	MemoryBuffers    uint64
	MemoryCached     uint64
	MemoryFree       uint64
	MemoryAvailable  uint64
	MemoryActive     uint64
	MemoryInactive   uint64
	MemorySwapTotal  uint64
	MemorySwapUsed   uint64
	MemorySwapCached uint64
	MemorySwapFree   uint64
	CreatedAt        time.Time
}

func GetMDADM(md string) DiskMDPerNode {
	out := DiskMDPerNode{}
	out.NodeID = Hostname()
	out.Kind = md

	cmd := exec.Command("mdadm", "-D", md)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			log.Panic(err)
		}
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	out.ExitCode = exitCode
	out.MDADM = fmt.Sprintf("%s%s", outStr, errStr)
	return out
}

func GetDiskUsage(path string) DiskUsagePerNode {
	disk := DiskUsagePerNode{Kind: path, NodeID: Hostname()}
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		log.Panic(err)
	}
	disk.DiskAll = fs.Blocks * uint64(fs.Bsize)
	disk.DiskFree = fs.Bfree * uint64(fs.Bsize)
	disk.DiskUsed = disk.DiskAll - disk.DiskFree

	return disk
}

func GetDiskIOStats(diskName string) DiskIOPerNode {
	out := DiskIOPerNode{}
	out.NodeID = Hostname()
	out.Kind = diskName

	disksStats, err := disk.Get()
	var diskIO *disk.Stats
	if err != nil {
		log.Panic(err)
	} else {
		for _, d := range disksStats {
			if d.Name == diskName {
				diskIO = &d
				break
			}
		}
		if diskIO == nil {
			log.Panicf("missing disk %s", diskName)
		}
	}

	out.DiskReadsCompleted = diskIO.ReadsCompleted
	out.DiskWritesCompleted = diskIO.WritesCompleted
	return out
}

func GetMemoryStats() MemStatsPerNode {
	out := MemStatsPerNode{}
	out.NodeID = Hostname()

	m, err := memory.Get()
	if err != nil {
		log.Panic(err)
	}

	return MemStatsPerNode{
		MemoryTotal:      m.Total,
		MemoryUsed:       m.Used,
		MemoryBuffers:    m.Buffers,
		MemoryCached:     m.Cached,
		MemoryFree:       m.Free,
		MemoryAvailable:  m.Available,
		MemoryActive:     m.Active,
		MemoryInactive:   m.Inactive,
		MemorySwapTotal:  m.SwapTotal,
		MemorySwapUsed:   m.SwapUsed,
		MemorySwapCached: m.SwapCached,
		MemorySwapFree:   m.SwapFree,
	}
}

func WriteStatsAndAlert(db *gorm.DB) error {

	//	d := GetDiskIOStats("md2")
	//	du := GetDiskUsageStats("/")
	//	m := GetMemoryStats()

	//	db.Save(n).Error
	return nil
}
