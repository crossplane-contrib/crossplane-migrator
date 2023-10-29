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

package main

import (
	"github.com/alecthomas/kong"

	"github.com/stevendborrelli/crossplane-migrator/newdeploymentconfig"
	"github.com/stevendborrelli/crossplane-migrator/newpipeinecomposition"
)

var _ = kong.Must(&cli)

var cli struct {
	NewDeploymentConfig    newdeploymentconfig.Cmd   `cmd:"" help:"Convert deprecated ControllerConfigs to DeploymentRuntimeConfigs."`
	NewPipelineComposition newpipeinecomposition.Cmd `cmd:"" help:"Convert Compositions to Composition Pipelines with function-patch-and-transform."`
}

func main() {
	//logger := logging.NewNopLogger()
	ctx := kong.Parse(&cli,
		kong.Name("crossplane-migrate"),
		kong.Description("Crossplane migration utilities"),
		// Binding a variable to kong context makes it available to all commands
		// at runtime.
		//kong.BindTo(logger, (logging.Logger)(nil)),
		//kong.Bind(logger),
		kong.ConfigureHelp(kong.HelpOptions{
			FlagsLast:      true,
			Compact:        true,
			WrapUpperBound: 80,
		}),
		kong.UsageOnError())
	ctx.FatalIfErrorf(ctx.Run())
}
