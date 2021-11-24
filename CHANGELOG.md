## [1.0.4] - 2021-11-24
### Added
* Support Kubernetes version 1.19-1.22 [#122](https://github.com/kuberlogic/kuberlogic/pull/122)


## [1.0.3] - 2021-11-22
### Added
* Add GKE support [#108](https://github.com/kuberlogic/kuberlogic/pull/108)
* Add simple installation script / Ask for a config parameters interactively when not found [#115](https://github.com/kuberlogic/kuberlogic/pull/115)
### Fixed / Improved
* Increase wait time for the loadbalancer check [#109](https://github.com/kuberlogic/kuberlogic/pull/109)
* Use versioned backup/restore images [#110](https://github.com/kuberlogic/kuberlogic/pull/110)
* Use post-install Helm hooks for managing release dependencies [#111](https://github.com/kuberlogic/kuberlogic/pull/111)
* Redirect users to the help page after Help button pressed [#112](https://github.com/kuberlogic/kuberlogic/pull/112)
* Use http communications between operator and grafana [#114](https://github.com/kuberlogic/kuberlogic/pull/114)
* Configure monitoring selectors via installer [#117](https://github.com/kuberlogic/kuberlogic/pull/117)

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
