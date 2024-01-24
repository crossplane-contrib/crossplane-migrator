
# crossplane-migrator <!-- omit in toc -->

**Note** this project is archived, as of https://github.com/crossplane/crossplane/pull/5275 and Crossplane 1.15 this functionality has been ported to the Crossplane CLI under `crossplane beta convert`.

---

Crossplane release [1.14](https://github.com/crossplane/crossplane/releases/tag/v1.14.0) introduced functions and the Deployment Runtime.

In order to take advantage of these features, this golang-based binary can be used to convert legacy YAML manifests to newer APIs.

The following conversions are supported:

- [ControllerConfig](https://docs.crossplane.io/latest/concepts/providers/#controller-configuration) to [DeploymentRuntimeConfig](https://docs.crossplane.io/latest/concepts/providers/#runtime-configuration)
- Compositions to use [Function-based Patch and Transform](https://github.com/crossplane-contrib/function-patch-and-transform) instead of the built-in engine.
  
This utility is proposed to be added to the Crossplane CLI in [1.15](https://github.com/crossplane/crossplane/issues/4922).

- [Installing](#installing)
- [Example Use](#example-use)
  - [Migrating ControllerConfigs to DeploymentRuntimeConfigs](#migrating-controllerconfigs-to-deploymentruntimeconfigs)
  - [Migrating Existing Compositions to use Pipelines](#migrating-existing-compositions-to-use-pipelines)
    - [Setting the Function Name](#setting-the-function-name)
    - [Migrating Running Compositions to Function Patch and Transform](#migrating-running-compositions-to-function-patch-and-transform)
- [Building](#building)
- [Known Issues](#known-issues)

## Installing

Binary releases are available at <https://github.com/crossplane-contrib/crossplane-migrator/releases>.

To install the latest version, run:

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

#### Migrating Running Compositions to Function Patch and Transform

*Note: please test this procedure in non-production environments and back up all resources before attempting a migration of live resources*.

First, ensure that a function that supports the legacy Path and Transform is installed in your cluster and is in a healthy state.

```shell
$ kubectl get function
NAME                           INSTALLED   HEALTHY   PACKAGE                                                       AGE
function-patch-and-transform   True        True      xpkg.upbound.io/upbound/function-patch-and-transform:v0.2.1   8m48s
```

Ensure all resources generated by this Composition are healthy:

```shell
$ kubectl get composite 
NAME              SYNCED   READY   COMPOSITION                         AGE
ref-aws-network   True     True    xnetworks.aws.platform.upbound.io   33m

```

Convert the composition and apply it to the cluster:

```shell
crossplane-migrator new-pipeline-composition -i composition.yaml -o composition-pt.yaml
```

```shell
$ kubectl apply -f composition-pt.yaml
composition.apiextensions.crossplane.io/xnetworks.aws.platform.upbound.io configured
```

Confirm that resources are healthy:

```shell
$ kubectl get composite 
NAME              SYNCED   READY   COMPOSITION                         AGE
ref-aws-network   True     True    xnetworks.aws.platform.upbound.io   33m

```

## Building

```console
go build -o crossplane-migrator
```

## Known Issues

- The migrator attempts to be as accurate as possible in mapping fields but has not been fully tested. The [ControllerConfig test suite](newdeploymentruntime/converter_test.go) and [Composition test suite](newpipelinecomposition/converter_test.go) attempt to cover all cases.
- The generated `DeploymentRuntimeConfig` has the same `Name:` as the ControllerConfig.
  
