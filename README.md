[![Build Status](https://travis-ci.com/kappnav/ui.svg?branch=master)](https://travis-ci.com/kappnav/operator)

# kAppNav Operator

Used to install Application Navigator. The operator is to read from a custom object that represents the application navigator instance. No "helm install" is needed.

## Build

To build the project run:

```
cd <kappnav project root>/operator 
./build.sh
```

This will compile the source within a Docker container and download required Go modules into the container at build time.

## Install and Uninstall

See [README](https://github.com/kappnav/README#install)

## Local Development

This project was developed with Operator SDK v0.10.0.

Installation instructions for the Operator SDK CLI are here:
https://github.com/operator-framework/operator-sdk/blob/master/doc/user/install-operator-sdk.md

Remember to replace the RELEASE_VERSION in the instructions with v0.10.0. It may not compile with newer versions.
Prerequisites for the Operator SDK are listed here: https://github.com/operator-framework/operator-sdk

If the structs for the Kappnav CRD (located in kappnav_types.go) are modified be sure to run:

```
operator-sdk generate k8s
operator-sdk generate openapi
```

This regenerates the CRD and the code that allows a Kappnav CR to be accessed programatically through the k8s APIs.

To build the project using the Operator SDK run:

```
cd <kappnav project root>/operator 
./buildWithSDK.sh
```

This will compile the source in your local environment using your local set of dependencies.

## Default values

Default values for the operator's configuration are stored in `deploy/default_values.yaml`. This CR file is included in the Docker image and is read each time a Kappnav CR is reconciled by the operator to fill in defaults for values that were not specified in the CR.

## Adding additional CRDs to the operator

Additional CRDs can be added to the `deploy/crds/extensions` folder. These will be included in the Docker image. The Application CRD is always included in the image. When the operator is installed it will attempt to create each of the CRDs in k8s if they do not already exist.

## Adding additional action, sections and status config maps to the operator

Additional action, sections and status config maps should be added to the `deploy/maps/action`, `deploy/maps/sections` and `deploy/maps/status` folders respectively. This supports the same templating language that is used in Helm charts. Variables are addressed by their field names in the Kappnav structs. For instance, the kubeEnv field from the CR would be addressed as `.Spec.Env.KubeEnv`. Action, sections and status config maps will be initially created when a CR is installed.

## Adding additional logic to the controller

If you are adding additional logic for managing resources, provide an implemenation of the `NewKappnavExtension` function in the `utils/extensions.go` file that returns an instance of `KappnavExtension`.

