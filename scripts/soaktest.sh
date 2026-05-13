#!/bin/bash
set -e

DURATION="${1:-3600}"  # seconds, default 1 hour
INTERVAL="${2:-10}"     # seconds between iterations
HEALTH_URL="${3:-http://localhost:1100/health}"

echo "=== Moistello Soak Test ==="
echo "Duration: ${DURATION}s"
echo "Interval: ${INTERVAL}s"
echo "Health URL: ${HEALTH_URL}"
echo ""

START=$(date +%s)
ITERATION=0
HEALTH_OK=0
HEALTH_FAIL=0

log_file="/tmp/moistello-soak-$(date +%Y%m%d-%H%M%S).log"
echo "Logging to: $log_file"

while true; do
    NOW=$(date +%s)
    ELAPSED=$((NOW - START))
    if [ $ELAPSED -ge $DURATION ]; then
        break
    fi

    ITERATION=$((ITERATION + 1))

    # Check health
    if curl -sf "$HEALTH_URL" > /dev/null 2>&1; then
        HEALTH_OK=$((HEALTH_OK + 1))
    else
        HEALTH_FAIL=$((HEALTH_FAIL + 1))
        echo "[$(date -Iseconds)] HEALTH FAIL" | tee -a "$log_file"
    fi

    # Progress
    if [ $((ITERATION % 10)) -eq 0 ]; then
        echo "[$(date -Iseconds)] iter=$ITERATION elapsed=${ELAPSED}s health_ok=${HEALTH_OK} health_fail=${HEALTH_FAIL}" | tee -a "$log_file"
    fi

    sleep "$INTERVAL"
done

echo ""
echo "=== Soak Test Complete ==="
echo "Duration: ${DURATION}s"
echo "Iterations: ${ITERATION}"
echo "Health OK: ${HEALTH_OK}"
echo "Health Fail: ${HEALTH_FAIL}"
echo "Log: $log_file"
