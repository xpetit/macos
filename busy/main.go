package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/getlantern/systray"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	. "github.com/xpetit/x"
)

func main() {
	go func() {
		for range time.Tick(5 * time.Second) {
			percent := Check2(cpu.Percent(4*time.Second, false))
			vm := Check2(mem.VirtualMemory())
			systray.SetTitle(fmt.Sprintf("%.f%% %.fG", percent[0]*float64(runtime.NumCPU()), float64(vm.Used)/1e9))
		}
	}()
	systray.Run(nil, nil)
}
