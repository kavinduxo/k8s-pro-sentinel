# K8s Pro Sentinel: Kubernetes Security Operator
<div align="center">
<img src="assets/img/pro-logo.png">
</div>

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
  - [Configuration](#configuration)
  - [Running Checks](#running-checks)
- [Documentation](#documentation)
  - [User Guide](#user-guide)
  - [Developer Guide](#developer-guide)
  - [API Reference](#api-reference)
- [Contributing](#contributing)
- [License](#license)

## Introduction

K8s Pro Sentinel is a powerful Kubernetes security operator designed to bolster the security of your Kubernetes clusters. It functions as a vigilant sentinel, continuously monitoring, auditing, and securing your cluster's resources, ensuring that your applications and infrastructure adhere to the highest security standards.

## Features

- **Automated Security Checks:** K8s Pro Sentinel automates a wide range of security checks, including RBAC policy analysis, secrets management, and more.

- **Comprehensive Auditing:** Gain deep insights into the security status of your cluster through detailed audit logs and reports. Easily track security violations and their resolution.

- **RBAC Best Practices:** Ensure that your Kubernetes Role-Based Access Control (RBAC) configurations adhere to best practices, minimizing the risk of unauthorized access.

- **Secrets Management:** Safeguard sensitive information by monitoring secrets usage and ensuring compliance with security policies.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).


### Prerequisites

Before getting started with K8s Pro Sentinel, ensure you have the following prerequisites:

- Kubernetes Cluster
- Kubectl (Kubernetes CLI) installed
- [Optional] Docker (for building custom images)

### Installation

1. Clone the K8s Pro Sentinel repository:

   ```sh
   git clone https://github.com/kavinduxo/k8s-pro-sentinel.git


### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/sentinal-operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/sentinal-operator:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

