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

package newpipeinecomposition

import (
	"errors"
	"fmt"
	"io"
	"os"

	"k8s.io/client-go/kubernetes/scheme"

	compositionv1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	v1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

// Cmd arguments and flags for render subcommand.
type Cmd struct {
	// Arguments.
	InputFile string `short:"i" type:"path" placeholder:"PATH" help:"The Composition file to be converted."`

	OutputFile   string `short:"o" type:"path" placeholder:"PATH" help:"The file to write the generated Composition to."`
	FunctionName string `short:"f" type:"string" placeholder:"PATH" help:"FunctionRef. Defaults to function-patch-and-transform."`
}

func (c *Cmd) Help() string {
	return `
This command converts a Crossplane Composition to use a Composition function pipeline.


Examples:

  # Write out a DeploymentRuntimeConfigFile from a ControllerConfig 

  migrator compositiontofunction -i composition.yaml -o new-composition.yaml

  # Use a different functionRef and output to stdout

  migrator compositiontofunction -i composition.yaml -f local-function-patch-and-transform

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
	_ = compositionv1.AddToScheme(sch)

	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode

	data, err := os.ReadFile(c.InputFile)
	if err != nil {
		return err
		//return errors.Errorf("unable to read ControllerConfig %s: %s", c.InputFile, err)
	}

	oc := &v1.Composition{}
	_, _, err = decode(data, &v1.CompositionGroupVersionKind, oc)
	if err != nil {
		fmt.Println("Decode error")
		return err
	}

	pc, err := NewPipelineCompositionFromExisting(oc, c.FunctionName)
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

	err = s.Encode(pc, output)
	if err != nil {
		return err
	}
	return nil

}
