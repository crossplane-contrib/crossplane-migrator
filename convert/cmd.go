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

package convert

import (
	"errors"
	"fmt"
	"io"
	"os"

	"k8s.io/client-go/kubernetes/scheme"

	"github.com/crossplane/crossplane/apis/pkg/v1alpha1"
	"github.com/crossplane/crossplane/apis/pkg/v1beta1"

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

  migrator convert -i my-controllerconfig.yaml -o my-drconfig.yaml

  # Create a new DeploymentRuntimeConfigFile via Stdout

  migrator convert -i cc.yaml | grep -v creationTimestamp | kubectl apply -f - 

`
}

func (c *Cmd) Run() error {
	if c.InputFile == "" {
		os.Stderr.Write([]byte(c.Help()))
		return errors.New("no input file")
	}

	// Set up schemes for our API types
	sch := runtime.NewScheme()
	_ = scheme.AddToScheme(sch)
	_ = v1alpha1.AddToScheme(sch)
	_ = v1beta1.AddToScheme(sch)

	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode

	data, err := os.ReadFile(c.InputFile)
	if err != nil {
		return err
		//return errors.Errorf("unable to read ControllerConfig %s: %s", c.InputFile, err)
	}

	cc := &v1alpha1.ControllerConfig{}
	_, _, err = decode(data, &v1alpha1.ControllerConfigGroupVersionKind, cc)
	if err != nil {
		fmt.Println("Decode error")
		return err
	}

	drc, err := ControllerConfigToRuntimeDeploymentConfig(cc)
	if err != nil {
		return err
	}

	s := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)

	var output io.Writer

	if c.OutputFile != "" {
		f, err := os.OpenFile(c.OutputFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		output = f
	} else {
		output = os.Stdout
	}

	err = s.Encode(drc, output)
	if err != nil {
		return err
	}
	return nil
}
