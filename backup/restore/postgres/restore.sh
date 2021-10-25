#! /usr/bin/env bash

# enable unofficial bash strict mode
set -o errexit
set -o nounset
set -o pipefail
IFS=$'\n\t'

# integrate sentry
SENTRY_CLI_NO_EXIT_TRAP=1
[[ -v SENTRY_DSN ]] && eval "$(sentry-cli bash-hook)"

PG_BIN=$PG_DIR/$PG_VERSION/bin

TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
K8S_API_URL=https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT/api/v1
CERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt

BACKUP_FILE_GZ=/tmp/backup.sql.gz
BACKUP_FILE=/tmp/backup.sql

function recreate_db() {
  # cannot using with "psql --single-transaction"
  $PG_BIN/psql \
    --echo-all \
    --set ON_ERROR_STOP=on \
    --set stringdb="'$DATABASE'" \
    --set db=$DATABASE \
    -f /app/recreatedb.sql
}

function extract_db() {
  # extract database if uses pg_dumpall for backup
  if egrep -n "PostgreSQL database cluster dump" $BACKUP_FILE; then
    DB_BACKUP=/tmp/$DATABASE.sql
    bash /app/pg_extract.sh $BACKUP_FILE $DATABASE $DB_BACKUP
    BACKUP_FILE=$DB_BACKUP
  fi
}

function restore() {
  $PG_BIN/psql \
    --echo-all \
    --single-transaction \
    --set ON_ERROR_STOP=on \
    --file $BACKUP_FILE \
    --dbname $DATABASE
}

function decompress() {
  gunzip $BACKUP_FILE_GZ
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

function get_master_pod() {
  get_pods "labelSelector=${CLUSTER_NAME_LABEL}%3D${SCOPE},spilo-role%3Dmaster" | head -n 1
}

CURRENT_NODENAME=$(get_current_pod | jq .items[].spec.nodeName --raw-output)
export CURRENT_NODENAME

for search in "${search_strategy[@]}"; do

  PGHOST=$(eval "$search")
  export PGHOST

  if [ -n "$PGHOST" ]; then
    break
  fi

done

set -x
aws_download
decompress
extract_db
recreate_db
restore
set +x

exit 0
