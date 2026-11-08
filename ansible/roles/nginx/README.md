# Nginx Role

## Overview

This Ansible role automates the installation, configuration, and deployment of Nginx on Ubuntu servers.

It includes configuration validation, conditional reloads, HTTP health checks, and automatic rollback when deployment fails.

The role is part of the **Server-Monitor** project, which uses a Rocky Linux Ansible controller to manage two Ubuntu servers.

## Features

- Install Nginx using APT.
- Start Nginx and enable it at boot.
- Create the website document root.
- Deploy a static HTML webpage.
- Deploy a server-specific Nginx configuration using Jinja2.
- Back up existing configurations before replacement.
- Validate configurations using `nginx -t`.
- Reload Nginx only when configuration changes.
- Check HTTP availability on port 8080.
- Attempt configuration rollback when deployment fails.
- Support rolling deployments through the main playbook.

## Role Structure

```text
nginx/
├── tasks/
│   └── main.yml
├── templates/
│   └── lab-nginx.conf.j2
├── files/
│   └── index.html
└── README.md
```

## Configuration

The role uses `lab-nginx.conf.j2` to generate server-specific configuration.

```nginx
server {
    listen 8080;
    server_tokens off;
    server_name {{ inventory_hostname }};

    root /var/www/html;
    index index.html;
}
```

The `inventory_hostname` variable allows each Ubuntu server to receive its corresponding hostname.

Nginx configuration is deployed to:

```text
/etc/nginx/conf.d/lab.conf
```

The HTML webpage is deployed to:

```text
/var/www/html/index.html
```

## Deployment Process

1. Install and start Nginx.
2. Ensure the document root exists and deploy the HTML page.
3. Check whether the existing Nginx configuration exists.
4. Deploy the Jinja2 template with backup enabled.
5. Validate the complete Nginx configuration.
6. Reload Nginx only if the configuration changed.
7. Perform an HTTP health check.
8. Attempt rollback if a deployment or health-check task fails.

### Rollback

The role uses Ansible `block` and `rescue` to handle deployment failures.

If validation or the health check fails, Ansible attempts to restore the previous configuration and reload Nginx.

The deployment is still reported as failed even when recovery succeeds.

**Limitations:** The role validates the configuration after replacing the live file. Rollback is not guaranteed for every failure, including unreachable hosts. HTTP 200 checks basic availability, not complete application health.

## Usage

Run from the `ansible/` directory.

### Syntax check

```bash
ansible-playbook site.yml --syntax-check
```

### Preview changes on web01

```bash
ansible-playbook site.yml --limit web01 --check --diff -K
```

### Deploy to web01

```bash
ansible-playbook site.yml --limit web01 -K
```

### Deploy to all web servers

```bash
ansible-playbook site.yml -K
```

The main playbook uses `serial: 1` to configure hosts one at a time.

## Testing

The following behaviors have been tested in the lab:

- Successful Nginx configuration deployment.
- Configuration validation.
- Conditional reload after configuration changes.
- Idempotency when no changes are required.
- Rollback after an intentionally invalid configuration on web01.
- HTTP availability checks.

## Requirements

- Ansible installed on the control node.
- Ubuntu managed hosts using APT and systemd.
- SSH connectivity to managed hosts.
- Sudo privileges for the SSH user.
- Nginx available from configured package repositories.

