#!/bin/bash
set -e

HEALTH_URL="${1:-http://localhost:1100/health}"
INTERVAL="${2:-5}"

echo "=== Moistello Health Monitor ==="
echo "Health URL: $HEALTH_URL"
echo "Interval: ${INTERVAL}s"
echo ""

LOG="/tmp/moistello-monitor-$(date +%Y%m%d-%H%M%S).log"

while true; do
    TIMESTAMP=$(date -Iseconds)

    if curl -sf "$HEALTH_URL" > /dev/null 2>&1; then
        echo "[$TIMESTAMP] ✓ HEALTHY" | tee -a "$LOG"
    else
        echo "[$TIMESTAMP] ✗ UNHEALTHY" | tee -a "$LOG"
    fi

    sleep "$INTERVAL"
done
