# KuberLogic
[![Development build](https://github.com/kuberlogic/kuberlogic/actions/workflows/on-push-master.yaml/badge.svg)](https://github.com/kuberlogic/kuberlogic/actions/workflows/on-push-master.yaml)
[![codecov](https://codecov.io/gh/kuberlogic/operator/branch/master/graph/badge.svg?token=VRWDPT0EIC)](https://codecov.io/gh/kuberlogic/operator)
---
![logo](img/kuberlogic-logo.png)

KuberLogic is a powerful way to deliver your application as-a-service.

KuberLogic is an open-source platform that helps to deliver any single-tenant application (one stack per customer) to multiple users as-a-service. It allows software vendors to accelerate their journey to Software-as-a-Service (SaaS).

## Features
* Tenant and application instance orchestration
* Automatic minor updates
* Automatic failover
* Scheduled backups
* REST API for service management

## Requirements
KuberLogic leverages a lot of top notch open-source projects and it requires a specific environment to run on top of:
* Kubernetes v1.20.x with:
  * StorageClass configured as a default
  * LoadBalancer Services
  * At least 2 nodes in cluster with 4G of RAM, 2 CPUs and 5G of disk space each
* S3 compatible storage for backups (optional)
* Helm v3.4 CLI tool

## Installation

1. Download `kuberlogic-installer` command-line installation tool for your platform from the [Releases page](https://github.com/kuberlogic/kuberlogic/releases) and add it to PATH.
2. Prepare the KuberLogic [configuration file](modules/installer/README.md#Configuration).
3. Run `kuberlogic-installer install all -c <configFile>`
4. Run `kuberlogic-installer status -c <configFile>`
5. Add DNS records for KuberLogic endpoints so they are pointing to KuberLogic Ingress IP. Alternatively, if you are evaluating KuberLogic, you may want to use /etc/hosts file to provide the access locally.

## Usage
### REST API
KuberLogic supports RESTful API for service management. To learn more visit:
* [API Scheme](https://editor.swagger.io/?url=https://raw.githubusercontent.com/kuberlogic/kuberlogic/master/modules/apiserver/openapi.yaml)


## Build KuberLogic
KuberLogic artifacts are container images and the installer binary. Requirements are:
* Golang 1.16
* Helm CLI 3.x
* Docker
### Build container images
`make docker-build docker-push`
### Build the installer
`make installer-build`
### Tests
KuberLogic includes a set of integration tests. They can be run in two modes:
* Local mode when KuberLogic operator & apiserver are started as goroutines:
* Remote mode when tests send API requests to a remote apiserver.

To run tests in both modes, dependencies must be installed into a Kubernetes cluster:
```shell
kuberlogic-installer install all
# setup Minio for backup & restore tests
cd modules/apiserver
make deploy-minio create-bucket
```
For more information see [docs](modules/apiserver/tests/README.md).

#### Local mode
```shell
cd modules/apiserver
make undeploy-operator generate-local-webhook-certs patch-endpoint
MY_VERSION=5.7.31 GODEBUG: x509ignoreCN=0 KUBERLOGIC_KUBECONFIGPATH=${HOME}/.kube/config make coverage-report RUN=/mysql 
```

#### Remote mode
```shell
cd modules/apiserver
MY_VERSION=5.7.31 GODEBUG: x509ignoreCN=0 KUBERLOGIC_KUBECONFIGPATH=${HOME}/.kube/config make remote-test REMOTE_HOST=<apiserver endpoint> RUN=/mysql
```

## Support
Feel free to open an [issue](https://github.com/kuberlogic/kuberlogic/issues) if you need any help. You can also reach us at info@kuberlogic.com with any questions. 

## License
```text
CloudLinux Software Inc 2019-2021 All Rights Reserved

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
