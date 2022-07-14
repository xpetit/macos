package main

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/getlantern/systray"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/xpetit/macos"
	. "github.com/xpetit/x"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const (
	maxMenuItems = 10
	refreshDelay = 5 * time.Second
)

func menuList(nb int) []*systray.MenuItem {
	m := make([]*systray.MenuItem, nb)
	for i := 0; i < nb; i++ {
		m[i] = systray.AddMenuItem("", "")
		m[i].Disable()
	}
	return m
}

func main() {
	systray.Run(func() {
		go func() {
			gigUsed := func() int {
				vm := Check2(mem.VirtualMemory())
				f := float64(vm.Used + vm.Wired)
				f /= 1e9         // B to GiB
				f /= 1.073741824 // GiB to GB
				return int(math.Ceil(f))
			}
			systray.SetTitle(fmt.Sprintf("??%% %dG", gigUsed()))
			for {
				t := time.Now()
				percent := Check2(cpu.Percent(refreshDelay-time.Second, false))
				systray.SetTitle(fmt.Sprintf("%.f%% %dG", percent[0]*float64(runtime.NumCPU()), gigUsed()))
				time.Sleep(time.Since(t.Add(refreshDelay)))
			}
		}()
		refreshButton := systray.AddMenuItem("Refresh", "Refresh list")

		systray.AddSeparator()

		memTotal := systray.AddMenuItem("", "")
		memTotal.Disable()

		systray.AddSeparator()

		mem := menuList(maxMenuItems)

		systray.AddSeparator()

		cpuTotal := systray.AddMenuItem("", "")
		cpuTotal.Disable()

		systray.AddSeparator()

		cpu := menuList(maxMenuItems)

		quit := systray.AddMenuItem("Quit", "")

		go func() {
			<-quit.ClickedCh
			systray.Quit()
		}()
		refresh := func() {
			processByName := macos.GetProcesses()
			var totalMem, totalCPU float64
			for _, p := range processByName {
				totalMem += float64(p.RSS)
				totalCPU += p.CPU
			}
			totalMem /= 1e6 // B to MB
			names := maps.Keys(processByName)
			slices.SortFunc(names, func(a, b string) bool {
				return processByName[a].RSS > processByName[b].RSS
			})
			memTotal.SetTitle(fmt.Sprintf("%.f MB Total memory used", totalMem))
			width := 1 + int(math.Log10(totalMem))
			for i, m := range mem {
				name := names[i]
				rss := float64(processByName[name].RSS) / 1e6 // B to MB
				m.SetTitle(fmt.Sprintf("%*.f MB  %3.f %%  %s", width, rss, 100*rss/totalMem, name))
			}

			cpuTotal.SetTitle(fmt.Sprintf("%5.1f %% Total CPU used", totalCPU))
			slices.SortFunc(names, func(a, b string) bool {
				return processByName[a].CPU > processByName[b].CPU
			})
			for i, m := range cpu {
				name := names[i]
				m.SetTitle(fmt.Sprintf("%5.1f %%  %s", processByName[name].CPU, name))
			}
		}
		refresh()
		go func() {
			for range refreshButton.ClickedCh {
				refresh()
			}
		}()
	}, nil)
}
