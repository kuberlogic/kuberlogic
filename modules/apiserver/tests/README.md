Installation:

```
# needed for the backup/restore tests
$ make run-minio 
```
Examples:
* all test with the single cluster instance -- use the short-test
```
$ make short-test
```

* test only the Service
```
$ make short-test TEST=TestService
```

* test only the Postgresql Service
```
$ make short-test TEST=TestService/postgresql
```