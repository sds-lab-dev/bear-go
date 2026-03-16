#!/usr/bin/env bash

start_dockerd() {
  mkdir -p /var/run /var/lib/docker

  cat >/etc/docker/daemon.json <<'EOF'
{
  "features": {
    "buildkit": true
  }
}
EOF

  dockerd --host=unix:///var/run/docker.sock > /var/log/dockerd.log 2>&1 &

  for i in $(seq 1 60); do
    if docker info >/dev/null 2>&1; then
      ready=1
      break
    fi
    sleep 1
  done

  if [ "${ready:-0}" -ne 1 ]; then
    echo "ERROR: dockerd did not become ready within 60 seconds" >&2
    exit 1
  fi
}
