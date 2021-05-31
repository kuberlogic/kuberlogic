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
``bash
make test REMOTE_DEPS=1 API_HOST=localhost API_PORT=10000 RUN=/postgresql
``