## [1.0.2] - 2021-11-09
### Added
* Kubernetes 1.21 support [#96](https://github.com/kuberlogic/kuberlogic/pull/96)
* Allow platform configuration for Kuberlogic components. [#106](https://github.com/kuberlogic/kuberlogic/pull/106)
* Add EKS support [#106](https://github.com/kuberlogic/kuberlogic/pull/106)
### Fixed
* Backup configurations can be edited again [#102](https://github.com/kuberlogic/kuberlogic/pull/102)
* Set authorization token lifetime to 1 hour [#101](https://github.com/kuberlogic/kuberlogic/pull/101)
* Backups now work with AWS S3 [#100](https://github.com/kuberlogic/kuberlogic/pull/100)
* Improved external address discovery for services [#99](https://github.com/kuberlogic/kuberlogic/pull/99)
* Use buildkit for building images [#92](https://github.com/kuberlogic/kuberlogic/pull/92)
* Validate Keycloak tokens using introspection endpoint [#92](https://github.com/kuberlogic/kuberlogic/pull/94)
* Service should return service name with namespace appended [#95](https://github.com/kuberlogic/kuberlogic/pull/95)
* Do not attempt to delete Grafana user if an org does not exit [#103](https://github.com/kuberlogic/kuberlogic/pull/103)

## [1.0.1] - 2021-10-11
### Added
* `kuberlogic-installer` is now built for darwin / linux: [#90](https://github.com/kuberlogic/kuberlogic/pull/90)

### Fixed
* `kuberlogic-installer` now shows monitoring console endpoint: [#91](https://github.com/kuberlogic/kuberlogic/pull/91)

## [1.0.0] - 2021-10-08
### Added
* Web UI
* Automatic provisioning
* Automatic minor updates
* Automatic failover
* Scheduled backups
* Resource usage monitoring
* Ability to scale up and down as needed
* REST API for service management
* Supported databases: postgresql (9.3 - 13) and mysql (5.7.24 - 5.7.31)

## [0.0.1] - 2021-09-08
### Added
- initial
