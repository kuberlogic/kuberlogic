#!/bin/bash
#
# CloudLinux Software Inc 2019-2021 All Rights Reserved
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

SLEEP_TIME=30
for i in $(seq 1 10); do
  kubectl get keycloakclient -o jsonpath='{.items[0].status.ready}' | grep true && exit 0
  sleep $SLEEP_TIME
done
echo "Keycloak is not ready in time!"
kubectl get keycloakclient -o yaml
kubectl logs keycloak-0
exit 1