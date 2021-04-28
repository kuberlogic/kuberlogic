#! /usr/bin/env bash

# enable unofficial bash strict mode
set -o errexit
set -o nounset
set -o pipefail
IFS=$'\n\t'

# integrate sentry
[[ -v SENTRY_DSN ]] && eval "$(sentry-cli bash-hook)"

TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
K8S_API_URL=https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT/api/v1
CERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt

BACKUP_FILE_GZ=/tmp/backup.sql.gz
BACKUP_FOLDER=/tmp/backup
RESULT=0

function restore() {
  # verbose=3 -- info
  myloader \
    --verbose 3 \
    --overwrite-tables \
    --enable-binlog \
    --host $MYSQL_HOST \
    --password $MYSQL_PASSWORD \
    --directory $BACKUP_FOLDER \
    --database $DATABASE
}

function decompress() {
  mkdir --parents $BACKUP_FOLDER
  tar --extract --verbose --directory $BACKUP_FOLDER --file $BACKUP_FILE_GZ
}

function exclude() {
  cd $BACKUP_FOLDER
  find . -not -name "$DATABASE*" -not -name "metadata" -delete
  ls -la $BACKUP_FOLDER
  cd -
}

function aws_download() {
  args=()

  [[ ! -z "$LOGICAL_BACKUP_S3_ENDPOINT" ]] && args+=("--endpoint-url=$LOGICAL_BACKUP_S3_ENDPOINT")
  [[ ! -z "$LOGICAL_BACKUP_S3_REGION" ]] && args+=("--region=$LOGICAL_BACKUP_S3_REGION")
  [[ ! -z "$LOGICAL_BACKUP_S3_SSE" ]] && args+=("--sse=$LOGICAL_BACKUP_S3_SSE")

  aws s3 cp $PATH_TO_BACKUP "$BACKUP_FILE_GZ" "${args[@]//\'/}"
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
  get_master_pod
)

function list_all_replica_pods_any_node() {
  get_pods "labelSelector=app.kubernetes.io/name%3Dmysql,role%3Dreplica,healthy%3Dyes,mysql.presslabs.org/cluster%3D${SCOPE}" | head -n 1
}

function get_master_pod() {
  get_pods "labelSelector=app.kubernetes.io/name%3Dmysql,role%3Dmaster,healthy%3Dyes,mysql.presslabs.org/cluster%3D${SCOPE}" | head -n 1
}

function check_replica_exists() {
  HOST=$(eval list_all_replica_pods_any_node)
  if [ -n "$HOST" ]; then
    echo "Replica exists $HOST. Should be available only the master node."
    exit 1
  fi
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
cd /tmp
aws_download && decompress && exclude && restore
RESULT=$?
cd -
set +x

exit $RESULT
