# kuberlogic-installer
Installer is responsible for managing Kuberlogic release - install, upgrade, uninstall processes. Its main technologies are Go, Helm SDK and Kubernetes' client-go library.

## Configuration
Configuration file must be prepared to run the installer. It is YAML formatted file and the default location is `$HOME/.kuberlogic-installer.yaml`

You can find a sample file with all the configuration references [here](sample-config.yaml).

## Installation
`kuberlogic-installer install`
Kuberlogic mixes together a lot of components, so the whole installation process is broke down to three phases:
* dependencies
* core
* platform
  You can specify any of those and there is also `all` phase that includes all of the above.

## Upgrade
`kuberlogic-installer upgrade`
Upgrade command will upgrade an existing Kuberlogic installation. The command will fail if a release is not found.
This command also supports phases just as the `install` command.

## Uninstall
`kuberlogic-installer uninstall`
Uninstall command will remove all of the Kuberlogic components except of:
* CRDs
* User-created services, backup configurations and restores

## Status
`kuberlogic-installer status`
`status` command shows current state of a Kuberlogic release and its connection endpoints. 