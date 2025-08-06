<div align="center">
<br>
<img src=".github/images/stackit-logo.svg" alt="STACKIT logo" width="50%"/>
<br>
<br>
</div>

# STACKIT Exporter
![Tests](https://github.com/h3adex/stackit-exporter/actions/workflows/test-lint.yaml/badge.svg)
![Vulnerability Scan](https://github.com/h3adex/stackit-exporter/actions/workflows/vuln-scan.yaml/badge.svg)
![Docker](https://github.com/h3adex/stackit-exporter/actions/workflows/build-release.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/h3adex/stackit-exporter)](https://goreportcard.com/report/github.com/h3adex/stackit-exporter)
![Go Version](https://img.shields.io/badge/go-1.24.5-blue)

> **Note:** This project is a community-driven initiative and is not officially associated with STACKIT. It is currently in the early stages of development. Contributions are welcome — feel free to open issues or submit pull requests to help improve the project.

## Documentation

Documentation for each collector is available in the [docs](docs) folder. Refer to it for detailed information on the exporter's functionality and setup.

## Installation

Container images are published via GitHub Packages. Example Kubernetes manifests can be found in the [k8s](./k8s) directory.

```shell
helm repo add stackit-exporter https://h3adex.github.io/stackit-exporter
helm repo update

helm install stackit-exporter ./stackit-exporter \
--namespace monitoring \
--create-namespace \
--set secret.name=stackit-exporter-secret \
--set serviceMonitor.label.release=prometheus-operator

helm install stackit-exporter ./stackit-exporter \
--namespace monitoring \
--create-namespace \
--set stackitCredentials.enabled=true \
--set stackitCredentials.projectID="XXX-XXX" \
--set stackitCredentials.serviceAccountKey="...." \
--set serviceMonitor.label.release=prometheus-operator
```

## Contributing

Contributions of all kinds are welcome! Whether it's bug reports, feature suggestions, or code improvements, your input is greatly appreciated.

If you'd like to add a new collector, please make sure to complete the following steps:

1. Add a new registry for the API in `/internal/metrics`.
2. Initialize the client and registry in `/cmd/main.go`.
3. Implement the collector and write tests for it in `/internal/collector`.
4. Write documentation in the `/docs` folder. See the [collector template](docs/collector.template.md) as a reference.
5. Ensure that lint checks and tests pass before opening a pull request.

## License

This project is licensed under the MIT License.