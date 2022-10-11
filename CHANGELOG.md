## [0.0.15] - 2022-10-11
- Support custom config files for managed applications
- Validate docker-compose file before installation
- Use ChargeBee features for managing applications feature based on subscriptions

## [0.0.14] - 2022-09-20
- Using "insecure" mode instead of tls-enabled. It means TLS enabled by default

## [0.0.13] - 2022-09-15
- Support custom domains when creating/editing services (`--domain` when using CLI and `domain` field when using REST API)
- Improve `kuberlogic install` command UX
- Check managed applications' readiness with HTTP based `ReadinessProbe`
- Minor bug fixes and stability improvements

## [0.0.12] - 2022-09-06
- Add 'info' command to CLI (check KuberLogic component statuses)
- Add 'version' command to CLI (check KuberLogic component versions)
- Add 'diag' command to CLI (gather information for debugging)
- Installation flow improvements (help texts, error messages, bug fixes)
- Add update-credentials capabilities for managed applications
- Various ChargeBee integration fixes
- Other minor bug fixes, stability improvements

## [0.0.11] - 2022-08-08
- Add 'install' command to KuberLogic CLI
- Add Sentry integration
- Update 'Service' 'model in KuberLogic REST API
- Support custom fields mapping in Chargebee integration
- More stability fixes

[0.0.14]: https://github.com/kuberlogic/kuberlogic/compare/0.0.13...0.0.14
[0.0.13]: https://github.com/kuberlogic/kuberlogic/compare/0.0.12...0.0.13
[0.0.12]: https://github.com/kuberlogic/kuberlogic/compare/0.0.11...0.0.12
[0.0.11]: https://github.com/kuberlogic/kuberlogic/releases/tag/0.0.11
