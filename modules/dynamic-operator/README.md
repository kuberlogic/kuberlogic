## How to install

```shell
# install crds
make install
# next command mainly needs for the webhook configuration, rbac install
make deploy
```

## How to run locally

```shell
# no need deployments for the local run
kubectl delete deploy dynamic-operator-controller-manager
# ability to work webhook for local run
make generate-local-webhook-certs patch-webhook-endpoint
cd /modules/dynamic-operator
# need to build plugin before running operator
go build -o plugin/postgres plugin/postgresql/operator.go
# export variable for the plugin enabling1
export PLUGINS="{postgresql,$(pwd)/plugin/postgres}"
make run # or go run main.go 
```


## How to run tests

```shell
cd /modules/dynamic-operator
go build -o plugin/postgres plugin/postgresql/operator.go
export PLUGINS="{postgresql,$(pwd)/plugin/postgres}"
make test
```