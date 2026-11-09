#!/bin/bash

disk_threshold=80

config_file="$HOME/server-monitor/config/monitor.conf"

health_failed=0
if [ -f "$config_file" ]; then
	source "$config_file"
	disk_threshold="$DISK_THRESHOLD"
fi

show_server(){
	echo "========== SERVER INFORMATION =========="
	echo
	echo "Server: $(hostname)"
	echo "Uptime: $(uptime -p)"
	echo "User: $(whoami)"
	echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
}

echo
show_memory() {
	total_ram=$(free -h | grep Mem | awk '{print $2}')
	available_ram=$(free -h | grep Mem | awk '{print $7}')

echo "========== MEMORY INFO =========="
echo
echo "Total RAM: $total_ram"
echo "Available RAM: $available_ram"
echo

}
echo
show_disk(){

	disk_usage=$(df -h / | awk 'NR==2 {print $5}')
	disk_usage=${disk_usage%\%}
	echo "Disk Usage: ${disk_usage}%"
echo

if [ "$disk_usage" -gt "$disk_threshold" ]; then
	echo "WARNING: Disk Usage is high!"
	health_failed=1
else
	echo  "Disk usage is normal."

fi
echo

}
echo
check_service(){
	local service="$1"

	if systemctl is-active --quiet "$service"; then
		echo "[OK] $service is running"
	else
		echo "[WARNING] $service is not running"
		health_failed=1
	fi

}

show_server
show_memory
show_disk

echo
echo
echo "========== SERVICE INFO =========="
services=("ssh" "nginx")
echo
for service in "${services[@]}"; do
	check_service "$service"
done

exit "$health_failed"
