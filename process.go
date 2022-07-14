package macos

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/xpetit/x"
)

type process struct {
	Command string
	CPU     float64
	RSS     int
	PID     int
	PPID    int
}

func parseTime(s string) (d time.Duration) {
	hours, s, ok := strings.Cut(s, ":")
	Assert("found hours", ok)
	d += time.Duration(Check2(strconv.Atoi(hours))) * time.Hour
	mins, s, ok := strings.Cut(s, ".")
	Assert("found minutes", ok)
	d += time.Duration(Check2(strconv.Atoi(mins))) * time.Minute
	d += time.Duration(Check2(strconv.Atoi(s))) * time.Second
	return
}

func parseElapsed(s string) (d time.Duration) {
	var ok bool
	if len(s) > 8 {
		var days string
		days, s, ok = strings.Cut(s, "-")
		Assert("days hours", ok)
		d += time.Duration(Check2(strconv.Atoi(days))) * 24 * time.Hour
	}
	if len(s) > 5 {
		var hours string
		hours, s, ok = strings.Cut(s, ":")
		Assert("found hours", ok)
		d += time.Duration(Check2(strconv.Atoi(hours))) * time.Hour
	}
	mins, s, ok := strings.Cut(s, ":")
	Assert("found minutes", ok)
	d += time.Duration(Check2(strconv.Atoi(mins))) * time.Minute
	d += time.Duration(Check2(strconv.Atoi(s))) * time.Second
	return
}

func GetProcesses() map[string]*process {
	// Parse command output
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
		CPUTime := parseTime(nextField())
		Uptime := parseElapsed(nextField())
		p := process{
			RSS:     Check2(strconv.Atoi(nextField())),
			PID:     Check2(strconv.Atoi(nextField())),
			PPID:    Check2(strconv.Atoi(nextField())),
			Command: strings.TrimSpace(line[i:]),
		}
		if Uptime > time.Minute {
			p.CPU = float64(CPUTime) / float64(Uptime)
		}
		if p.PID > 1 {
			processByPID[p.PID] = &p
		}
	}

	// Add to the top level process (the parent process immediately below PID 1) the resources of each of its children
	for _, p := range processByPID {
		if p.PPID == 1 {
			continue
		}
		parent := p
		for ; parent.PPID != 1; parent = processByPID[parent.PPID] {
		}
		parent.CPU += p.CPU
		parent.RSS += p.RSS
	}

	// Merge processes that are under the same name/category
	processByName := map[string]*process{}
	for _, p := range processByPID {
		if p.PPID != 1 {
			continue
		}
		name := p.Command
		switch name {
		case "/System/Library/PrivateFrameworks/XprotectFramework.framework/Versions/A/XPCServices/XprotectService.xpc/Contents/MacOS/XprotectService":
			name = "XProtect"
		default:
			switch {
			case
				strings.HasPrefix(name, "/Applications/JavaSnipt"),
				strings.HasPrefix(name, "/Applications/StopTheMadness"),
				strings.HasPrefix(name, "/Applications/Wipr"),
				strings.HasPrefix(name, "/Library/Apple/System/Library/CoreServices/SafariSupport.bundle"),
				strings.HasPrefix(name, "/System/Library/Frameworks/AppKit.framework"),
				strings.HasPrefix(name, "/System/Library/Frameworks/SafariServices.framework"),
				strings.HasPrefix(name, "/System/Library/Frameworks/WebKit.framework"),
				strings.HasPrefix(name, "/System/Library/PrivateFrameworks/SafariSafeBrowsing.framework"),
				strings.HasPrefix(name, "/System/Library/PrivateFrameworks/SafariShared.framework"),
				strings.HasSuffix(name, "com.giorgiocalderolla.Wipr-Mac.Wipr-Refresher"):
				name = "/Applications/Safari"
			default:
				if i := strings.Index(name, ".app/Contents/"); i != -1 {
					name = name[:i]
				}
			}
		}
		name = filepath.Base(name)
		if p2, ok := processByName[name]; !ok {
			// insert a copy
			p := *p
			p.Command = name
			processByName[name] = &p
		} else {
			// merge
			p2.CPU += p.CPU
			p2.RSS += p.RSS
		}
	}
	for _, p := range processByName {
		p.RSS *= 1024
		p.CPU = p.CPU * 5 / 3
	}
	return processByName
}
