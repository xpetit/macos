package main

import (
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/xpetit/x"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type process struct {
	command string
	CPUTime time.Duration
	Uptime  time.Duration
	RSS     int
	PID     int
	PPID    int
}

func parseDuration(s, minuteSep string) (d time.Duration) {
	if len(s) > 5 {
		hours, s2, ok := strings.Cut(s, ":")
		Assert("found hours", ok)
		d += time.Duration(Check2(strconv.Atoi(hours))) * time.Hour
		s = s2
	}
	mins, s, ok := strings.Cut(s, minuteSep)
	Assert("found minutes", ok)
	d += time.Duration(Check2(strconv.Atoi(mins))) * time.Minute
	d += time.Duration(Check2(strconv.Atoi(s))) * time.Second
	return
}

func main() {
	// parse command output
	s := string(Check2(exec.Command("ps", "-ASo", "time,etime,rss,pid,ppid,comm").Output()))
	lines := strings.Split(s, "\n")
	lines = lines[1 : len(lines)-1] // remove header and trailing line
	processByPID := map[int]*process{}
	for _, line := range lines {
		var i int
		nextField := func() string {
			for ; line[i] == ' '; i++ {
			}
			start := i
			for ; line[i] != ' '; i++ {
			}
			return line[start:i]
		}
		p := process{
			CPUTime: parseDuration(nextField(), "." /*minuteSep*/),
			Uptime:  parseDuration(nextField(), ":" /*minuteSep*/),
			RSS:     Check2(strconv.Atoi(nextField())),
			PID:     Check2(strconv.Atoi(nextField())),
			PPID:    Check2(strconv.Atoi(nextField())),
			command: strings.TrimSpace(line[i:]),
		}
		if p.PID > 1 {
			processByPID[p.PID] = &p
		}
	}

	// for each process, add its memory usage (RSS) to the ancestor (the parent process immediately below PID 1)
	memByPID := map[int]int{}
	for _, p := range processByPID {
		rss := p.RSS
		for ; p.PPID != 1; p = processByPID[p.PPID] {
		}
		memByPID[p.PID] += rss
	}

	// merge processes that are under the same name/category
	memByCmd := map[string]int{}
	var total int
	for pid, mem := range memByPID {
		if processByPID[pid].PPID == 1 {
			mem *= 1024
			total += mem
			cmd := processByPID[pid].command
			switch cmd {
			case "/System/Library/PrivateFrameworks/XprotectFramework.framework/Versions/A/XPCServices/XprotectService.xpc/Contents/MacOS/XprotectService":
				cmd = "XProtect"
			default:
				switch {
				case
					strings.HasPrefix(cmd, "/Applications/JavaSnipt"),
					strings.HasPrefix(cmd, "/Applications/StopTheMadness"),
					strings.HasPrefix(cmd, "/Applications/Wipr"),
					strings.HasPrefix(cmd, "/Library/Apple/System/Library/CoreServices/SafariSupport.bundle"),
					strings.HasPrefix(cmd, "/System/Library/Frameworks/AppKit.framework"),
					strings.HasPrefix(cmd, "/System/Library/Frameworks/SafariServices.framework"),
					strings.HasPrefix(cmd, "/System/Library/Frameworks/WebKit.framework"),
					strings.HasPrefix(cmd, "/System/Library/PrivateFrameworks/SafariSafeBrowsing.framework"),
					strings.HasPrefix(cmd, "/System/Library/PrivateFrameworks/SafariShared.framework"),
					strings.HasSuffix(cmd, "com.giorgiocalderolla.Wipr-Mac.Wipr-Refresher"):
					cmd = "/Applications/Safari"
				default:
					if i := strings.Index(cmd, ".app/Contents/"); i != -1 {
						cmd = cmd[:i]
					}
				}
			}
			cmd = filepath.Base(cmd)
			memByCmd[cmd] += mem
		}
	}

	// print results
	cmds := maps.Keys(memByCmd)
	slices.SortFunc(cmds, func(a, b string) bool {
		return memByCmd[a] > memByCmd[b]
	})
	tot := fmt.Sprintf("%.f", float64(total)/1e6)
	fmt.Println(tot, "MB         Total memory used")
	fmt.Println()
	for _, cmd := range cmds {
		percent := strconv.Itoa(int(math.Round(100 * float64(memByCmd[cmd]) / float64(total))))
		fmt.Printf("%*.f MB  %3s %%  %s \n", len(tot), float64(memByCmd[cmd])/1e6, percent, cmd)
	}
}
