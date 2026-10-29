#!/bin/bash

server=("sasan1@192.168.56.103" "sasan2@192.168.56.102")

log_file="$HOME/server-monitor/logs/central_monitor.log"


check_server(){
	server="$1"
	echo "====== Checking Server ====== "
	if ssh -o ConnectTimeout=5 "$server" "bash ~/server-monitor/scripts/system_info.sh"; then
		echo "[OK] Monitoring completed: $server"
	else
		echo "[ERROR] Failed to monitor: $server"
	fi
echo

}


echo "====== MONITORING STARTED ======"
echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
echo

for server in "${server[@]}"; do
	check_server "$server"
done

echo "COMPLETED"
