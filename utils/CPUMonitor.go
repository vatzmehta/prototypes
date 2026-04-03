package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

func StartCPUMonitor(done <-chan struct{}) *[]float64 {
	readings := &[]float64{}
	cpu.Percent(0, false) // establish baseline
	var mu sync.Mutex
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				pct, err := cpu.Percent(0, false)
				if err == nil && len(pct) > 0 {
					mu.Lock()
					*readings = append(*readings, pct[0])
					mu.Unlock()
				}
			}
		}
	}()
	return readings
}

func PrintCPUGraph(readings []float64) {
	const barWidth = 50
	fmt.Println("\nCPU Usage over time (sampled every 1s):")
	fmt.Println(strings.Repeat("-", barWidth+18))
	for i, pct := range readings {
		filled := int(pct / 100.0 * barWidth)
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		fmt.Printf("%3ds | %s %.1f%%\n", i+1, bar, pct)
	}
	fmt.Println(strings.Repeat("-", barWidth+18))
}
