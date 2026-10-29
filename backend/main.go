package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func runCommand(command string, args ...string) (string, error) {
	output, err := exec.Command(command, args...).CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.Trimspace(string(output)))
	}
	return strings.Trimspace(string(output)), nil
}

func main() {

	output, err := runCommand(
		"bash",
		"/home/sasan/server-monitor/scripts/centralmonitor.sh",
	)

	if err != nill {
		fmt.Println("Monitoring failed:", err)
		return

	}
	fmt.Println(output)
}
