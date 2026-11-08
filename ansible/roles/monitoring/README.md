# Monitoring Role

## Overview

This Ansible role deploys and manages a Bash-based Linux monitoring script and its configuration on Ubuntu servers.

It is part of the **Server-Monitor** project, which uses Rocky Linux as an Ansible controller to manage and monitor two Ubuntu servers.

The role manages monitoring files. Actual monitoring is performed by `system_info.sh`, which is executed remotely through SSH by `centralmonitor.sh`.

## Features

- Create monitoring directories on managed servers.
- Deploy the existing `system_info.sh` Bash script.
- Generate per-server monitoring configurations using Jinja2.
- Configure disk usage warning thresholds through Ansible host variables.
- Maintain executable permissions for monitoring scripts.
- Support idempotent deployments.

## Role Structure

```text
monitoring/
├── tasks/
│   └── main.yml
├── files/
│   └── system_info.sh
├── templates/
│   └── monitor.conf.j2
└── README.md
```

## Monitoring Architecture

```text
Rocky Linux Controller
│
├── Ansible
│   └── Deploys monitoring role
│
└── centralmonitor.sh
    │
    ├── SSH → web01 → system_info.sh
    │
    └── SSH → web02 → system_info.sh
```

Ansible handles deployment and configuration management.

`centralmonitor.sh` independently connects to the Ubuntu servers through SSH to execute their monitoring scripts and retrieve results.

## Monitored Information

The deployed `system_info.sh` collects:

- Hostname and system uptime.
- Current user and date.
- Total and available memory.
- Root filesystem disk usage.
- Disk usage threshold warnings.
- SSH and Nginx service status.

## Configuration

The Jinja2 configuration template, `monitor.conf.j2`, contains:

```jinja2
DISK_THRESHOLD={{ disk_threshold }}
```

The `disk_threshold` value is supplied by Ansible host variables.

For example:

```yaml
disk_threshold: 80
```

After rendering, the deployed configuration contains:

```bash
DISK_THRESHOLD=80
```

This allows servers to use different disk warning thresholds without modifying the Bash script.

## Deployment Paths

For the project's current Ubuntu home-directory layout:

```text
/home/<ssh-user>/server-monitor/
├── config/
│   └── monitor.conf
├── scripts/
│   └── system_info.sh
└── logs/
```

The SSH usernames are `sasan1` on web01 and `sasan2` on web02.

The deployment role is designed around these users' existing home-directory structure.

## Usage

Run from the `ansible/` directory.

### Check syntax

```bash
ansible-playbook monitoring.yml --syntax-check
```

### Preview deployment

```bash
ansible-playbook monitoring.yml --limit web01 --check --diff -K
```

### Deploy monitoring on web01

```bash
ansible-playbook monitoring.yml --limit web01 -K
```

### Deploy to both servers

```bash
ansible-playbook monitoring.yml -K
```

## Run Monitoring

After deployment, execute the collector on the Rocky Linux controller:

```bash
cd ~/server-monitor
bash scripts/centralmonitor.sh
```

This connects to both Ubuntu servers and runs their deployed `system_info.sh` scripts.

The collector works independently of Ansible once deployment is complete.

## Testing

The monitoring role has been tested for:

- Successful deployment to both Ubuntu servers.
- Rendering host-specific disk thresholds.
- Correct file and directory paths.
- Idempotency of deployed monitoring files.
- SSH-based execution through `centralmonitor.sh`.
- Collection of memory, disk, uptime, and service information.

## Requirements

- Ansible on the control node.
- SSH access to managed Ubuntu servers.
- Bash and common Linux utilities.
- systemd for service checks.
- Valid `disk_threshold` values in Ansible host variables.
- Write access to the managed user's monitoring directories.

## Current Limitations

- Monitoring is executed on demand, not continuously.
- Monitoring results are currently human-readable.
- A successful script exit does not necessarily mean every monitored service is healthy.
- Centralized structured log collection and alerting are not yet implemented.

