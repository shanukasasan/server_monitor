package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"bufio"
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

type HistoryEntry struct {
	Timestamp time.Time     `json:"timestamp"`
	Servers []ServerStatus  `json:"servers"`

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

var (
	latestResults []ServerStatus
	latestError	error
	lastUpdated	time.Time
	mu		sync.RWMutex
)

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

func collectServers() ([]ServerStatus, error) {
    inventory, err := getInventory()
    if err != nil {
        return nil, err
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
            result.Status = "ERROR"
            result.Error = "Invalid monitoring JSON: " + err.Error()
        }

        result.Hostname = hostname
        results = append(results, result)
    }

    return results, nil
}

func startMonitoring() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    collect := func() {
        results, err := collectServers()
	collectedAt := time.Now()

        mu.Lock()
        if err == nil {
            latestResults = results
        }
        latestError = err
        lastUpdated = time.Now()
        mu.Unlock()

        if err != nil {
            log.Println("Automatic monitoring failed:", err)
	return
        } 
	if err := saveHistory(results, collectedAt); err != nil {
		log.Println("Failed to save monitoring history:", err)
	}

	log.Println("Automatic monitoring completed")
    }

    collect()

    for range ticker.C {
        collect()
    }
}



func monitorHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    results, err := collectServers()

    if err != nil {
        log.Println("Monitoring error:", err)

        w.WriteHeader(http.StatusInternalServerError)

        json.NewEncoder(w).Encode(map[string]string{
            "error": "Failed to collect server information",
        })
        return
    }

    json.NewEncoder(w).Encode(MonitorResponse{
        Servers: results,
    })
}

func serversHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    mu.RLock()

    results := append([]ServerStatus(nil), latestResults...)
    updated := lastUpdated
    collectionErr := latestError

    mu.RUnlock()

    if updated.IsZero() || collectionErr != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Latest monitoring collection is unavailable",
        })
        return
    }

    json.NewEncoder(w).Encode(MonitorResponse{
        Servers: results,
    })
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    file, err := os.Open(
        "/home/sasan/server-monitor/logs/history.jsonl",
    )
    if err != nil {
        http.Error(w, "Failed to open monitoring history",
            http.StatusInternalServerError)
        return
    }
    defer file.Close()

    history := []HistoryEntry{}

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        var entry HistoryEntry

        if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
            log.Println("Invalid history entry:", err)
            continue
        }

        history = append(history, entry)
    }

    if err := scanner.Err(); err != nil {
        log.Println("History read error:", err)
        http.Error(w, "Failed to read monitoring history",
            http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(history)
}

func saveHistory(results []ServerStatus, timestamp time.Time) error {
    file, err := os.OpenFile(
        "/home/sasan/server-monitor/logs/history.jsonl",
        os.O_APPEND|os.O_CREATE|os.O_WRONLY,
        0600,
    )
    if err != nil {
        return err
    }
    defer file.Close()

    entry := HistoryEntry{
        Timestamp: timestamp,
        Servers:   results,
    }

    return json.NewEncoder(file).Encode(entry)
}


func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/monitor", monitorHandler)
	http.HandleFunc("/servers", serversHandler)
	http.HandleFunc("/history", historyHandler)
	http.Handle("/", http.FileServer(http.Dir("backend/static")))

	go startMonitoring()

	log.Println("Server-Monitor API listening on port 8081")

	log.Fatal(http.ListenAndServe("127.0.0.1:8081", nil))
}
