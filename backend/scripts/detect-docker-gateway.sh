#!/bin/bash
# Detects the default-bridge (docker0) gateway IP of the docker daemon the
# judge uses, and writes it into config.yaml (docker.hostGatewayIP).
#
# Judge-spawned containers reach services published on their docker host
# (e.g. SonarQube) via this address — it is host-specific (daemon.json bip /
# address-pool overrides, subnet-conflict avoidance), so it must be queried
# from the actual daemon instead of hardcoded. Honors docker.mode in
# config.yaml: "remote" queries protocol://host:port, anything else queries
# the local daemon.
set -euo pipefail

CONFIG_FILE="config.yaml"

[ -f "$CONFIG_FILE" ] || { echo "$CONFIG_FILE not found (run from backend root)"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "docker CLI is required"; exit 1; }

# Reads a scalar under the top-level docker:/sonar: block — plain grep would
# also match keys of the same name in other sections (e.g. database.*.host).
yaml_val() {
  local section="$1" key="$2"
  awk -v section="$section" -v key="$key" '
    $0 ~ "^"section":" {in_section=1; next}
    in_section && /^[^[:space:]#]/ {in_section=0}
    in_section && $1 == key":" {gsub(/["'"'"']/, "", $2); print $2; exit}
  ' "$CONFIG_FILE"
}

mode=$(yaml_val docker mode)
if [ "$mode" = "remote" ]; then
  protocol=$(yaml_val docker protocol)
  host=$(yaml_val docker host)
  port=$(yaml_val docker port)
  export DOCKER_HOST="${protocol:-tcp}://${host}:${port}"
  echo "docker.mode=remote — querying daemon at $DOCKER_HOST"
else
  echo "docker.mode=${mode:-local} — querying local daemon"
fi

GATEWAY_IP=$(docker network inspect bridge --format '{{(index .IPAM.Config 0).Gateway}}')
[ -n "$GATEWAY_IP" ] || { echo "Failed to detect the default-bridge gateway IP"; exit 1; }
echo "Default-bridge (docker0) gateway: $GATEWAY_IP"

if grep -q "hostGatewayIP:" "$CONFIG_FILE"; then
  sed -i "s|hostGatewayIP: .*|hostGatewayIP: \"$GATEWAY_IP\"|" "$CONFIG_FILE"
else
  sed -i "/^docker:/a\\  hostGatewayIP: \"$GATEWAY_IP\"" "$CONFIG_FILE"
fi
echo "docker.hostGatewayIP updated in $CONFIG_FILE"

# Bonus: verify SonarQube is reachable from a container on this daemon —
# the exact network path grading containers will use.
sonar_port=$(yaml_val sonar port)
sonar_url="http://$GATEWAY_IP:${sonar_port:-9000}/api/system/status"
echo "Verifying SonarQube reachability from a container: $sonar_url"
if status=$(docker run --rm curlimages/curl -s -m 5 "$sonar_url" 2>/dev/null) && echo "$status" | grep -q '"status":"UP"'; then
  echo "SonarQube is reachable from containers: $status"
else
  echo "WARNING: SonarQube is NOT reachable at $sonar_url from a container." >&2
  echo "         Check that docker-compose.sonar.yml publishes \"$GATEWAY_IP:${sonar_port:-9000}:9000\" on this host." >&2
fi
