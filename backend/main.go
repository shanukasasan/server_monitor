package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

func runCommand(command string, args ...string) (string, error) {
	output, err := exec.Command(command, args...).CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}

	return strings.TrimSpace(string(output)), nil
}

type ServerStatus struct {
	Hostname  string `json:"hostname"`
	DiskUsage int    `json:"disk_usage"`
	SSH       string `json:"ssh"`
	Nginx     string `json:"nginx"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

type MonitorResponse struct {
	Servers []ServerStatus `json:"servers"`
}

type Inventory struct {
	Webservers struct {
		Hosts []string `json:"hosts"`
	} `json:"webservers"`

	Meta struct {
		HostVars map[string]struct {
			Address string `json:"ansible_host"`
			User    string `json:"ansible_user"`
			Port    int    `json:"ansible_port"`
		} `json:"hostvars"`
	} `json:"_meta"`
}

func getInventory() (Inventory, error) {
	var inventory Inventory

	output, err := runCommand(
		"ansible-inventory",
		"-i",
		"/home/sasan/server-monitor/ansible/inventory.ini",
		"--list",
	)

	if err != nil {
		return inventory, err
	}

	err = json.Unmarshal([]byte(output), &inventory)
	return inventory, err
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Server-Monitor API is running",
	})
}

func monitorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	inventory, err := getInventory()

	if err != nil {
		http.Error(w, `{"error":"Failed to read Ansible inventory"}`,
			http.StatusInternalServerError)
		log.Println("Inventory error:", err)
		return
	}

	results := []ServerStatus{}

	for _, hostname := range inventory.Webservers.Hosts {
		host := inventory.Meta.HostVars[hostname]

		address := host.Address
		if address == "" {
			address = hostname
		}

		result := ServerStatus{
			Hostname: hostname,
		}

		if host.User == "" {
			result.Status = "ERROR"
			result.Error = "ansible_user is not configured"
			results = append(results, result)
			continue
		}

		port := host.Port
		if port == 0 {
			port = 22
		}

		target := host.User + "@" + address

		output, err := runCommand(
			"ssh",
			"-o", "BatchMode=yes",
			"-o", "ConnectTimeout=5",
			"-p", strconv.Itoa(port),
			target,
			"bash ~/server-monitor/scripts/system_info.sh --json",
		)

		if err != nil {
			result.Status = "ERROR"
			result.Error = err.Error()
		} else if err := json.Unmarshal([]byte(output), &result); err != nil {
			result.Hostname = hostname
			result.Status = "ERROR"
			result.Error = "Invalid monitoring JSON: " + err.Error()
		} else {

			result.Hostname = hostname
		}

		results = append(results, result)
	}

	json.NewEncoder(w).Encode(MonitorResponse{
		Servers: results,
	})
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/monitor", monitorHandler)

	log.Println("Server-Monitor API listening on port 8081")

	log.Fatal(http.ListenAndServe("127.0.0.1:8081", nil))
}
