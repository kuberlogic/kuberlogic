# KuberLogic
[![Operator tests](https://github.com/kuberlogic/kuberlogic/actions/workflows/operator.yaml/badge.svg)](https://github.com/kuberlogic/kuberlogic/actions/workflows/operator.yaml)
[![Apiserver unittests](https://github.com/kuberlogic/kuberlogic/actions/workflows/apiserver.yaml/badge.svg)](https://github.com/kuberlogic/kuberlogic/actions/workflows/apiserver.yaml)
[![codecov](https://codecov.io/gh/kuberlogic/kuberlogic/master/graph/badge.svg?token=VRWDPT0EIC)](https://codecov.io/gh/kuberlogic/kuberlogic)
---
![logo](img/kuberlogic-logo.png)

----

KuberLogic is an open-source solution that helps to deliver any single-tenant application (one stack per customer) to multiple users as-a-cloud service. KuberLogic allows software vendors to accelerate their journey to Software-as-a-Service (SaaS) with minimal modifications to the application.

### Installation

Follow [Installing KuberLogic](https://kuberlogic.com/docs/getting-started) to set up your environment and install KuberLogic.

### Features

- Application instance (Tenant) orchestration (list/provision/delete)
- Scheduled and Instant backups
- Custom domain (subdomain) support
- SSL support
- Integration with billing provider (Chargebee) via webhooks
- Application (Tenant) isolation
- Application instance updates
- RESTful API and CLI for service management
- Integration with Sentry for errors and exceptions tracking

### Coming soon

[You can check our Roadmap here](https://kuberlogic.clearflask.com/roadmap)

----

### Why use KuberLogic?

The ultimate goal of KuberLogic is to provide an easily accessible service to turn any containerized application into a cloud-native SaaS solution.

KuberLogic:

- Provides a straightforward and reliable way to deploy and manage application instances (Tenants) while achieving maximum resource utilization and standardization;
- Simplify migration to multi-tenancy using industry-standard containers & K8s and allows rapid migration to SaaS with minimal application modification;
- Gives independence and frees from vendor lock, as KuberLogic is open source and based on Kubernetes to provide a consistent platform anywhere

### Requirements

Kubernetes cluster 1.20-1.23

### Documentation

Please refer to the official docs at [kuberlogic.com](https://docs.kuberlogic.com/)

### Getting involved

Feel free to open an [issue](https://github.com/kuberlogic/kuberlogic/issues) if you need any help.

You can see the roadmap/announcements and leave your feedback [here](https://kuberlogic.clearflask.com/).

You can also reach out to us at [info@kuberlogic.com](mailto:info@kuberlogic.com)
 or join our [Slack community](https://join.slack.com/t/kuberlogic/shared_invite/zt-x845lggh-lne0taYmwLFgQ6XZEiTJoA)

## License
```text
CloudLinux Software Inc 2019-2022 All Rights Reserved

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
