package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/xpetit/x"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type process struct {
	Command string
	RSS     int
	PID     int
	PPID    int
}

func main() {
	// Parse command output
	s := string(Check2(exec.Command("ps", "-ASo", "rss,pid,ppid,comm").Output()))
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
			RSS:     Check2(strconv.Atoi(nextField())),
			PID:     Check2(strconv.Atoi(nextField())),
			PPID:    Check2(strconv.Atoi(nextField())),
			Command: strings.TrimSpace(line[i:]),
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
			p2.RSS += p.RSS
		}
	}

	var totalMem int
	for _, p := range processByName {
		p.RSS *= 1024
		totalMem += p.RSS
	}

	// Display results
	names := maps.Keys(processByName)
	slices.SortFunc(names, func(a, b string) bool {
		return processByName[a].RSS > processByName[b].RSS
	})
	tot := fmt.Sprintf("%.f", float64(totalMem)/1e6)
	fmt.Println(tot, "MB Total memory used")
	fmt.Println()
	for _, cmd := range names {
		percent := fmt.Sprintf("%.f", 100*float64(processByName[cmd].RSS)/float64(totalMem))
		fmt.Printf("%*.f MB %3s %% %s \n", len(tot), float64(processByName[cmd].RSS)/1e6, percent, cmd)
	}
}
