#! /usr/bin/env bash

# enable unofficial bash strict mode
set -o errexit
set -o nounset
set -o pipefail
IFS=$'\n\t'

BACKUP_NAME=$(date '+%s')
[[ ! -z "$DATABASE" ]] && BACKUP_NAME=$DATABASE-$BACKUP_NAME

BACKUP_FILE=${BACKUP_NAME}.tar.gz
RESULT=0

TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
K8S_API_URL=https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT/api/v1
CERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt

function dump() {
  mkdir --parents $BACKUP_NAME

  args=()
  [[ ! -z "$DATABASE" ]] && args+=("--database=$DATABASE")

  # verbose=3 -- info
  mydumper \
    --build-empty-files \
    --triggers \
    --routines \
    --events \
    --verbose 3 \
    --host $MYSQL_HOST \
    --password $MYSQL_PASSWORD \
    --outputdir $BACKUP_NAME \
    "${args[@]//\'/}"
}

function compress() {
  cd $BACKUP_NAME
  tar -czvf ../$BACKUP_FILE .
  cd -
}

function aws_upload() {
  PATH_TO_BACKUP=s3://$LOGICAL_BACKUP_S3_BUCKET"/mysql/"$SCOPE$LOGICAL_BACKUP_S3_BUCKET_SCOPE_SUFFIX"/logical_backups/"$BACKUP_FILE

  args=()

  [[ ! -z "$LOGICAL_BACKUP_S3_ENDPOINT" ]] && args+=("--endpoint-url=$LOGICAL_BACKUP_S3_ENDPOINT")
  [[ ! -z "$LOGICAL_BACKUP_S3_REGION" ]] && args+=("--region=$LOGICAL_BACKUP_S3_REGION")
  [[ ! -z "$LOGICAL_BACKUP_S3_SSE" ]] && args+=("--sse=$LOGICAL_BACKUP_S3_SSE")

  aws s3 cp $BACKUP_FILE "$PATH_TO_BACKUP" "${args[@]//\'/}"
}

function get_pods() {
  declare -r SELECTOR="$1"

  curl "${K8S_API_URL}/namespaces/${POD_NAMESPACE}/pods?$SELECTOR" \
    --cacert $CERT \
    -H "Authorization: Bearer ${TOKEN}" | jq .items[].status.podIP -r
}

function get_current_pod() {
  curl "${K8S_API_URL}/namespaces/${POD_NAMESPACE}/pods?fieldSelector=metadata.name%3D${HOSTNAME}" \
    --cacert $CERT \
    -H "Authorization: Bearer ${TOKEN}"
}

declare -a search_strategy=(
  list_all_replica_pods_current_node
  list_all_replica_pods_any_node
  get_master_pod
)

function list_all_replica_pods_current_node() {
  get_pods "labelSelector=app.kubernetes.io/name%3Dmysql,role%3Dreplica,healthy%3Dyes,mysql.presslabs.org/cluster%3D${SCOPE}&fieldSelector=spec.nodeName%3D${CURRENT_NODENAME}" | head -n 1
}

function list_all_replica_pods_any_node() {
  get_pods "labelSelector=app.kubernetes.io/name%3Dmysql,role%3Dreplica,healthy%3Dyes,mysql.presslabs.org/cluster%3D${SCOPE}" | head -n 1
}

function get_master_pod() {
  get_pods "labelSelector=app.kubernetes.io/name%3Dmysql,role%3Dmaster,healthy%3Dyes,mysql.presslabs.org/cluster%3D${SCOPE}" | head -n 1
}

CURRENT_NODENAME=$(get_current_pod | jq .items[].spec.nodeName --raw-output)
export CURRENT_NODENAME

for search in "${search_strategy[@]}"; do

  MYSQL_HOST=$(eval "$search")
  export MYSQL_HOST

  if [ -n "$MYSQL_HOST" ]; then
    break
  fi

done

set -x
dump && compress && aws_upload
RESULT=$?
set +x

exit $RESULT
