package main

import (
	"fmt"
	"log"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)

func runCommand(command string, args ...string) (string, error) {
	output, err := exec.Command(command, args...).CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{

	"status": "ok",
	"message": "Server-Monitor API is running",

	})

}

func monitorHandler(w http.ResponseWriter, r *http.Request) {

	output, err := runCommand(
		"bash",
		"/home/sasan/server-monitor/scripts/centralmonitor.sh",
	)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"message": err.Error(),

		})
		return

	}

	json.NewEncoder(w).Encode(map[string]string{

		"status": "completed",
		"output": output,

	})

}


func main() {

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/monitor", monitorHandler)

	log.Println("Server-Monitor API listening on port 8081")
	log.Fatal(http.ListenAndServe("127.0.0.1:8081", nil))

}
