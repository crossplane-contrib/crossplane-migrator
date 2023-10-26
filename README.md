# ControllerConfig-migrator

Migrates deprecated Crossplane ControllerConfig to DeploymentRuntimeConfig

This command converts a Crossplane ControllerConfig to a DeploymentRuntimeConfig.

DeploymentRuntimeConfig was introduced in Crossplane 1.14 and ControllerConfig is
deprecated.

## Example Use

```console
  # Write out a DeploymentRuntimeConfigFile from a ControllerConfig 

  migrator convert -i my-controllerconfig.yaml -o my-drconfig.yaml

  # Create a new DeploymentRuntimeConfigFile via Stdout

  migrator convert -i cc.yaml | grep -v creationTimestamp | kubectl apply -f - 
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

## Building

```console
go build -o migrator
```

## Known Issues

- Output `metadata` fields contain a `creationTimestamp`
- The migrator attempts to be as accurate as possible in mapping the fields but has not been fully tested. The [test_suite](convert/converter_test.go) validates conformance
  