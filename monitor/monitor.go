package monitor

import (
	"log/slog"
	"sort"
	"time"

	"ai-monitoring/config"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
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

// GetUsage CPU와 메모리 사용률만 가볍게 조회합니다.
func GetUsage() (float64, float64, error) {
	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		return 0, 0, err
	}
	cpuUsage := 0.0
	if len(cpuPercents) > 0 {
		cpuUsage = cpuPercents[0]
	}

	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, err
	}
	return cpuUsage, v.UsedPercent, nil
}

// GetSystemState 현재 시스템 상태 반환
func GetSystemState(topN int) (*SystemState, error) {
	cpuUsage, memUsage, err := GetUsage()
	if err != nil {
		return nil, err
	}

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
		// 무거운 프로세스 정보 수집은 임계치 초과 시에만 수행해 부하를 줄입니다.
		cpuUsage, memUsage, err := GetUsage()
		if err != nil {
			slog.Error("시스템 상태를 가져오는데 실패했습니다", "error", err)
			continue
		}

		slog.Debug("현재 상태", "cpu", cpuUsage, "mem", memUsage)

		if cpuUsage > cfg.CPUThreshold || memUsage > cfg.MemoryThreshold {
			slog.Warn("임계치 초과 감지!", "cpu", cpuUsage, "mem", memUsage)

			state, err := GetSystemState(10)
			if err != nil {
				slog.Error("프로세스 정보 수집 실패", "error", err)
				continue
			}
			triggerCh <- state
		}
	}
}
