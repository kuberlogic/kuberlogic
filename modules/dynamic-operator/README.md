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
kubectl delete deploy kls-controller-manager
# ability to work webhook for local run
make generate-local-webhook-certs patch-webhook-endpoint
cd /modules/dynamic-operator
# need to build plugin before running operator
make build-docker-compose-plugin
# export variable for the plugin enabling1
export PLUGINS="{postgresql,$(pwd)/bin/docker-compose-plugin}"
make run # or go run main.go 
```


## How to run tests

```shell
cd /modules/dynamic-operator
make build-docker-compose-plugin
export PLUGINS="{postgresql,$(pwd)/bin/docker-compose-plugin}"
make test
```