[![integration tests](https://github.com/kuberlogic/kuberlogic/actions/workflows/test.yaml/badge.svg?branch=master)](https://github.com/kuberlogic/kuberlogic/actions/workflows/test.yaml)
[![codecov](https://codecov.io/gh/kuberlogic/operator/branch/master/graph/badge.svg?token=VRWDPT0EIC)](https://codecov.io/gh/kuberlogic/operator)

## Kuberlogic
With KuberLogic you can deploy, manage and scale software using Kubernetes in a frictionless way. Any supported software is turned into managed SaaS - hosted on your Kubernetes cluster. We have API, monitoring, backups, and integration with SSO right out of the box.

### Requirements
Kuberlogic leverages a lot of top notch open-source projects and it requires a specific environment to run on top of:
* Kubernetes v1.20.x with:
   * StorageClass configured as a default
   * LoadBalancer Services
   * At least 3 nodes in cluster with 4G of RAM and 2 CPUs each
* S3 compatible storage for backups (optional)
* Helm v3.4 CLI tool

### Installation

1. Download `kuberlogic-installer` command-line installation tool
2. Prepare the configuration file.
3. Run `kuberlogic-installer install all -c <configFile>`
4. Access Kuberlogic Web UI via connection endpoint from the previous step.
