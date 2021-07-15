## Installation

For the backup & restore tests you need to configure storage:
```bash
make deploy-minio 
make create-bucket 
```

If you need to know backup credentials you can execute

```bash
make show-minio-credentials
```

## Examples

* all test with the single cluster instance

```bash
make test
```

* run only the Service tests (for all services: postgresql, mysql, etc.)

```bash
make test RUN=TestLogs
```

* run all tests for Postgresql

```bash
make test RUN=/postgresql
```

* run only the Postgresql DB tests

```bash
make test RUN=TestDb/postgresql
```

* run tests against remote Kuberlogic installation
```bash
make test REMOTE_HOST=localhost:8080
```

## Run tests inside cluster
First please make sure that the necessary version of the operator is deployed and running
```bash
make deploy-minio 
make create-bucket 
export RUN=TestDb/postgresql # if you need to run a specific test, not required
make test-in-cluster # create a k8s job with tests
```

to show test's logs in follow mode:
```bash
make watch-test-in-cluster 
```

to remove test's artifacts such as kuberlogicservices, kuberlogictenants, existing test's job:
*it removes all kuberlogicservices (for all namespaces), all kuberlogictenants and all jobs*
```bash
make clear-test-in-cluster
```