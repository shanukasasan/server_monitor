#!/bin/bash

server=("sasan1@192.168.56.103" "sasan2@192.168.56.102")

log_file="$HOME/server-monitor/logs/central_monitor.log"


check_server(){
	local server="$1"
	local exit_code
	echo "====== Checking Server ====== "
	ssh -o ConnectTimeout=5 "$server" "bash ~/server-monitor/scripts/system_info.sh"


	exit_code=$?
	case "$exit_code" in
		0)
		   echo "[HEALTHY] $server"
		   ;;
		1)
		   echo "[UNHEALTHY] $server"
		   ;;

		255)
		   echo "[UNREACHABLE] $server"
		   ;;

		*)
		   echo "[ERROR] Monitoring execution failed on $server (exit code: $exit_code)"
		   ;;
	esac
echo

}

mkdir -p "$(dirname "$logfile")"

{
	echo "====== MONITORING STARTED ======"
	echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
	echo

	for server in "${server[@]}"; do
		check_server "$server"
	done

	echo "COMPLETED"
	echo
} 2>&1 | tee -a "$log_file"
