package www

import (
	"os/exec"
	"runtime"
)

// Browse opens the specified URL in the default browser
func Browse(url string, browser string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", browser}
	case "darwin":
		cmd = "open"
		if browser != "" {
			args = []string{"-a", browser}
		}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
