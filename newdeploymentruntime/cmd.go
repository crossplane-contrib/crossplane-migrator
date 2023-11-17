/*
Copyright 2023 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package newdeploymentruntime

import (
	"io"
	"os"

	"k8s.io/client-go/kubernetes/scheme"

	"github.com/crossplane/crossplane/apis/pkg/v1alpha1"
	"github.com/crossplane/crossplane/apis/pkg/v1beta1"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

// Cmd arguments and flags for render subcommand.
type Cmd struct {
	// Arguments.
	InputFile string `short:"i" type:"path" placeholder:"PATH" help:"The ControllerConfig file to be Converted."`

	OutputFile string `short:"o" type:"path" placeholder:"PATH" help:"The file to write the generated DeploymentRuntimeConfig to."`
}

func (c *Cmd) Help() string {
	return `
This command converts a Crossplane ControllerConfig to a DeploymentRuntimeConfig.

DeploymentRuntimeConfig was introduced in Crossplane 1.14 and ControllerConfig is
deprecated.

Examples:

  # Write out a DeploymentRuntimeConfigFile from a ControllerConfig 

  crossplane-migrator new-deployment-runtime -i examples/enable-flags.yaml -o my-drconfig.yaml

  # Create a new DeploymentRuntimeConfigFile via Stdout

  crossplane-migrator new-deployment-runtime -i cc.yaml | grep -v creationTimestamp | kubectl apply -f - 

`
}

func (c *Cmd) Run(logger logging.Logger) error {
	var data []byte
	var err error

	if c.InputFile != "" {
		data, err = os.ReadFile(c.InputFile)
	} else {
		data, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		return errors.Wrap(err, "Unable to read input")
	}

	// Set up schemes for our API types
	sch := runtime.NewScheme()
	_ = scheme.AddToScheme(sch)
	_ = v1alpha1.AddToScheme(sch)
	_ = v1beta1.AddToScheme(sch)

	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode

	cc := &v1alpha1.ControllerConfig{}
	_, _, err = decode(data, &v1alpha1.ControllerConfigGroupVersionKind, cc)
	if err != nil {
		return errors.Wrap(err, "Decode Error")
	}
	if cc.Spec.ServiceAccountName != nil && *cc.Spec.ServiceAccountName != "" {
		logger.Info("WARNING: serviceAccountName is set in the ControllerConfig.\nDeploymentRuntime does not create serviceAccounts, please create the service account separately.", "serviceAccountName", *cc.Spec.ServiceAccountName)
	}
	drc, err := ControllerConfigToDeploymentRuntimeConfig(cc)
	if err != nil {
		return errors.Wrap(err, "Cannot migrate to Deployment Runtime")
	}

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

	var output io.Writer

	if c.OutputFile != "" {
		f, err := os.OpenFile(c.OutputFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errors.Wrap(err, "Unable to open output file")
		}
		defer f.Close()
		output = f
	} else {
		output = os.Stdout
	}

	err = s.Encode(drc, output)
	if err != nil {
		return errors.Wrap(err, "Unable to encode output")
	}
	return nil
}
