# crossplane-migrator <!-- omit in toc -->

This golang-based binary migrates YAML manifests to newer APIs. The following migrations are supported:

- ControllerConfig to DeploymentRuntimeConfig
- Compositions to use Function-based Patch&Transform instead of the built-in engine.
  
This utility is proposed to be added to the Crossplane CLI in [1.15](https://github.com/crossplane/crossplane/issues/4922).

- [Installing](#installing)
- [Example Use](#example-use)
  - [Migrating ControllerConfigs to DeploymentRuntimeConfigs](#migrating-controllerconfigs-to-deploymentruntimeconfigs)
  - [Migrating Existing Compositions to use Pipelines](#migrating-existing-compositions-to-use-pipelines)
    - [Setting the Function Name](#setting-the-function-name)
- [Building](#building)
- [Known Issues](#known-issues)

## Installing

```shell
go install github.com/crossplane-contrib/crossplane-migrator@latest
```

## Example Use

```console
Usage: crossplane-migrator <command>

Crossplane migration utilities

Commands:
  new-deployment-runtime      Convert deprecated ControllerConfigs to
                              DeploymentRuntimeConfigs.
  new-pipeline-composition    Convert Compositions to Composition Pipelines with
                              function-patch-and-transform.
```

### Migrating ControllerConfigs to DeploymentRuntimeConfigs

[Crossplane ControllerConfig](https://docs.crossplane.io/latest/concepts/packages/#speccontrollerconfigref) to a [DeploymentRuntimeConfig](https://github.com/crossplane/crossplane/blob/master/design/one-pager-package-runtime-config.md).

DeploymentRuntimeConfig was introduced in Crossplane 1.14 and ControllerConfig has been marked [deprecated](https://github.com/crossplane/crossplane/issues/3601) since Crossplane 1.11.0.

Write out a DeploymentRuntimeConfig file from a ControllerConfig manifest:

```console
crossplane-migrator new-deployment-runtime  -i examples/enable-flags.yaml  -o my-drconfig.yaml
```

Create a new DeploymentRuntimeConfig via Stdout

```console
crossplane-migrator new-deployment-runtime -i cc.yaml | grep -v creationTimestamp | kubectl apply -f - 
```  

Once the new `DeploymentRuntimeConfig` has been created on the Crossplane Cluster it can be used by Function and Provider
packages via the `runtimeConfigRef`, which replaces `controllerConfigRef`.

```yaml
apiVersion: pkg.crossplane.io/v1beta1
kind: Function
metadata:
  name: function-patch-and-transform
spec:
  package: xpkg.upbound.io/crossplane-contrib/function-patch-and-transform:v0.1.4
  runtimeConfigRef:
    apiVersion: pkg.crossplane.io/v1beta1   # currently apiVersion and kind are optional
    kind: DeploymentRuntimeConfig
    name: func-env

```

### Migrating Existing Compositions to use Pipelines

[function-patch-and-transform](https://github.com/crossplane-contrib/function-patch-and-transform) runs Crossplane's built-in Patch&Transform engine in a function pipeline. This utility can migrate existing compositions to run in a pipeline:

```console
./crossplane-migrator new-pipeline-composition -i composition.yaml -o composition-pipeline.yaml
```

#### Setting the Function Name

By default, the `functionNameRef` is set to `function-patch-and-transform`. If installing a function via a Crossplane Configuration package, the package organization will be added (i.e. `crossplane-contrib-function-patch-and-transform`).

```shell
./crossplane-migrator new-pipeline-composition --function-name crossplane-contrib-function-patch-and-transform -i composition.yaml
```

## Building

```console
go build -o crossplane-migrator
```

## Known Issues

- The migrator attempts to be as accurate as possible in mapping the fields but has not been fully tested. The [ControllerConfig test suite](newdeploymentruntime/converter_test.go) [Composition test suite](newpipeinecomposition/converter_test.go) attempt to cover all cases.
- The generated `DeploymentRuntimeConfig` has the same `Name:` as the ControllerConfig
- Output `metadata` fields contain a `creationTimestamp`. This is a known Kubernetes issue that may be addressed via PR <https://github.com/kubernetes/kubernetes/pull/120757> merged in October 2023. Until upstream tooling is updated, remove the field from manifests.
  