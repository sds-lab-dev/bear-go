#!/usr/bin/env bash

start_dockerd() {
  local pidfile=/var/run/docker.pid
  local sock=/var/run/docker.sock
  local log=/var/log/dockerd.log
  local unix_sock="unix://$sock"
  local pid=""
  local exe=""
  local ready=0

  mkdir -p /var/run /var/lib/docker /var/log /etc/docker

  cat >/etc/docker/daemon.json <<'EOF'
{
  "features": {
    "buildkit": true
  }
}
EOF

  # Does the Docker PID file exist?
  if [ -f "$pidfile" ]; then
    pid="$(cat "$pidfile" 2>/dev/null || true)"

    # Is a process that has the PID alive?
    if [ -z "$pid" ] || [ ! -d "/proc/$pid" ]; then
      echo "Removing stale docker pid file: $pidfile" >&2
      rm -f "$pidfile" "$sock"
    else
      # Get the process name.
      exe="$(readlink -f "/proc/$pid/exe" 2>/dev/null || true)"
      # Is the process docker daemon?
      if [[ "$exe" == *"/dockerd" ]]; then
        echo "Existing dockerd found: pid=$pid exe=$exe" >&2
        echo "Stopping existing dockerd (pid=$pid)" >&2
        kill -TERM "$pid" 2>/dev/null || true

        for _ in $(seq 1 20); do
          [ ! -d "/proc/$pid" ] && break
          sleep 1
        done

        if [ -d "/proc/$pid" ]; then
          echo "dockerd did not stop gracefully; sending SIGKILL (pid=$pid)" >&2
          kill -KILL "$pid" 2>/dev/null || true
        fi        
      else
        echo "ERROR: pidfile points to non-dockerd process: pid=$pid exe=$exe" >&2
        exit 1
      fi      
    fi
  fi

  # Clear stale files.
  rm -f "$pidfile" "$sock"

  dockerd --host="$unix_sock" >"$log" 2>&1 &
  local dockerd_pid=$!

  for _ in $(seq 1 60); do
    if docker info >/dev/null 2>&1; then
      ready=1
      break
    fi

    if [ ! -d "/proc/$dockerd_pid" ]; then
      echo "ERROR: dockerd exited early. See $log" >&2
      tail -100 "$log" >&2 || true
      exit 1
    fi

    sleep 1
  done

  if [ "$ready" -ne 1 ]; then
    echo "ERROR: dockerd did not become ready within time limit" >&2
    tail -100 "$log" >&2 || true
    exit 1
  fi
}