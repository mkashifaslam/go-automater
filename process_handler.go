package main

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	targetVolume = "/Volumes/ExternalSSD"
)

var whitelist = regexp.MustCompile(`(?i)(node|npm|pnpm|yarn|code|Code|IntelliJ|idea|java)`)

type Process struct {
	PID     int    `json:"pid"`
	Command string `json:"command"`
	User    string `json:"user"`
	Action  string `json:"action"`
}

type Result struct {
	DryRun           bool      `json:"dryRun"`
	Processes        []Process `json:"processes"`
	UnmountAttempted bool      `json:"unmountAttempted"`
	UnmountSuccess   bool      `json:"unmountSuccess"`
	ForceEjectUsed   bool      `json:"forceEjectUsed"`
	Message          string    `json:"message"`
}

// run performs the main logic of finding processes and attempting unmount
func run(w http.ResponseWriter, dryRun bool) {
	user := currentUser()
	pids := findPIDsUsingVolume()

	var processes []Process

	for _, pid := range pids {
		cmd, owner := processInfo(pid)
		if owner != user {
			continue
		}

		action := "skip"
		if whitelist.MatchString(cmd) {
			action = "terminate"
			if !dryRun {
				_ = exec.Command("kill", "-TERM", strconv.Itoa(pid)).Run()
			}
		}

		processes = append(processes, Process{
			PID:     pid,
			Command: cmd,
			User:    owner,
			Action:  action,
		})
	}

	sort.Slice(processes, func(i, j int) bool {
		return processes[i].PID < processes[j].PID
	})

	result := Result{
		DryRun:           dryRun,
		Processes:        processes,
		UnmountAttempted: !dryRun,
		UnmountSuccess:   false,
		ForceEjectUsed:   false,
	}

	// --- Auto-unmount logic ---
	if !dryRun {
		result.UnmountAttempted = true

		// Try normal unmount
		if err := unmountVolume(false); err == nil {
			result.UnmountSuccess = true
			result.Message = "Volume unmounted successfully"
		} else {
			// Fallback: force eject
			if err := unmountVolume(true); err == nil {
				result.UnmountSuccess = true
				result.ForceEjectUsed = true
				result.Message = "Force eject was required but succeeded"
			} else {
				result.Message = "Unmount failed even after force eject"
			}
		}
	} else {
		result.Message = "Dry run only — no unmount attempted"
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// findPIDsUsingVolume uses lsof to find PIDs using the target volume
func findPIDsUsingVolume() []int {
	out, err := exec.Command("lsof").Output()
	if err != nil {
		return nil
	}

	pidSet := map[int]struct{}{}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		if strings.HasPrefix(fields[8], targetVolume) {
			pid, err := strconv.Atoi(fields[1])
			if err == nil {
				pidSet[pid] = struct{}{}
			}
		}
	}

	var pids []int
	for pid := range pidSet {
		pids = append(pids, pid)
	}

	return pids
}

// processInfo retrieves the command name and user for a given PID
func processInfo(pid int) (command, user string) {
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=,user=").Output()
	if err != nil {
		return "", ""
	}

	parts := strings.Fields(string(out))
	if len(parts) >= 2 {
		command = parts[0]
		user = parts[1]
	}
	return
}

// unmountVolume attempts to unmount the target volume
func unmountVolume(force bool) error {
	args := []string{"unmount"}
	if force {
		args = append(args, "force")
	}
	args = append(args, targetVolume)

	cmd := exec.Command("diskutil", args...)
	return cmd.Run()
}

// currentUser retrieves the current system user
func currentUser() string {
	out, err := exec.Command("whoami").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
