#!/bin/bash
SLEEP_TIME=30
for i in $(seq 1 10); do
  kubectl get keycloakclient -o jsonpath='{.items[0].status.ready}' | grep true && exit 0
done
echo "Keycloak is not ready in time!"
exit 1