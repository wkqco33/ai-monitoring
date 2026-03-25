package monitor

import (
	"log/slog"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
	"ai-monitoring/config"
)

type ProcessInfo struct {
	PID     int32
	Name    string
	CPUProg float64
	MemProg float32
}

type SystemState struct {
	Timestamp time.Time
	CPUUsage  float64
	MemUsage  float64
	Processes []ProcessInfo
}

// GetSystemState 현재 시스템 상태 반환
func GetSystemState(topN int) (*SystemState, error) {
	// CPU
	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	cpuUsage := 0.0
	if len(cpuPercents) > 0 {
		cpuUsage = cpuPercents[0]
	}

	// Memory
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	memUsage := v.UsedPercent

	// Processes
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var pInfos []ProcessInfo
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		cpuP, _ := p.CPUPercent()
		memP, _ := p.MemoryPercent()

		if cpuP > 0.1 || memP > 0.1 {
			pInfos = append(pInfos, ProcessInfo{
				PID:     p.Pid,
				Name:    name,
				CPUProg: cpuP,
				MemProg: memP,
			})
		}
	}

	// Sort by CPU usage
	sort.Slice(pInfos, func(i, j int) bool {
		return pInfos[i].CPUProg > pInfos[j].CPUProg
	})

	if len(pInfos) > topN {
		pInfos = pInfos[:topN]
	}

	return &SystemState{
		Timestamp: time.Now(),
		CPUUsage:  cpuUsage,
		MemUsage:  memUsage,
		Processes: pInfos,
	}, nil
}

// Start 모니터링 루프 시작
func Start(cfg *config.AppConfig, triggerCh chan<- *SystemState) {
	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		state, err := GetSystemState(10) // Top 10 프로세스
		if err != nil {
			slog.Error("시스템 상태를 가져오는데 실패했습니다", "error", err)
			continue
		}

		slog.Debug("현재 상태", "cpu", state.CPUUsage, "mem", state.MemUsage)

		if state.CPUUsage > cfg.CPUThreshold || state.MemUsage > cfg.MemoryThreshold {
			slog.Warn("임계치 초과 감지!", "cpu", state.CPUUsage, "mem", state.MemUsage)
			triggerCh <- state
		}
	}
}
