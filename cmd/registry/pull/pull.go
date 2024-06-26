// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2023 The Diginfra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pull

import (
	"context"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/diginfra/diginfractl/internal/utils"
	ociutils "github.com/diginfra/diginfractl/pkg/oci/utils"
	"github.com/diginfra/diginfractl/pkg/options"
	"github.com/diginfra/diginfractl/pkg/output"
)

const (
	longPull = `Pull Diginfra "rulesfile" or "plugin" OCI artifacts from remote registry.

Artifact references are passed as arguments.

A reference is a fully qualified reference ("<registry>/<repository>"),
optionally followed by ":<tag>" (":latest" is assumed by default when no tag is given).

Example - Pull artifact "myplugin" for the platform where diginfractl is running (default) in the current working directory (default):
	diginfractl registry pull localhost:5000/myplugin:latest

Example - Pull artifact "myplugin" for platform "linux/arm64" in the current working directory (default):
	diginfractl registry pull localhost:5000/myplugin:latest --platform linux/arm64

Example - Pull artifact "myplugin" for platform "linux/arm64" in "myDir" directory:
	diginfractl registry pull localhost:5000/myplugin:latest --platform linux/arm64 --dest-dir=./myDir

Example - Pull artifact "myrulesfile":
	diginfractl registry pull localhost:5000/myrulesfile:latest
`
)

type pullOptions struct {
	*options.Common
	*options.Artifact
	*options.Registry
	destDir string
}

func (o *pullOptions) Validate() error {
	return o.Artifact.Validate()
}

// NewPullCmd returns the pull command.
func NewPullCmd(ctx context.Context, opt *options.Common) *cobra.Command {
	o := pullOptions{
		Common:   opt,
		Artifact: &options.Artifact{},
		Registry: &options.Registry{},
	}

	cmd := &cobra.Command{
		Use:                   "pull hostname/repo[:tag|@digest] [flags]",
		DisableFlagsInUseLine: true,
		Short:                 "Pull a Diginfra OCI artifact from remote registry",
		Long:                  longPull,
		Args:                  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Validate(); err != nil {
				return err
			}

			ref := args[0]

			_, err := utils.GetRegistryFromRef(ref)
			if err != nil {
				return err
			}
			o.Common.Initialize()
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.RunPull(ctx, args)
		},
	}

	o.Registry.AddFlags(cmd)
	output.ExitOnErr(o.Printer, o.Artifact.AddFlags(cmd))
	cmd.Flags().StringVarP(&o.destDir, "dest-dir", "o", "", "destination dir where to save the artifacts(default: current directory)")
	return cmd
}

// RunPull executes the business logic for the pull command.
func (o *pullOptions) RunPull(ctx context.Context, args []string) error {
	logger := o.Printer.Logger
	ref := args[0]

	registry, err := utils.GetRegistryFromRef(ref)
	if err != nil {
		return err
	}

	puller, err := ociutils.Puller(o.PlainHTTP, o.Printer)
	if err != nil {
		return fmt.Errorf("an error occurred while creating the puller for registry %s: %w", registry, err)
	}

	err = ociutils.CheckConnectionForRegistry(ctx, puller.Client, o.PlainHTTP, registry)
	if err != nil {
		return err
	}

	logger.Info("Preparing to pull artifact", logger.Args("name", args[0]))

	if o.destDir == "" {
		logger.Info("Pulling artifact in the current directory")
	} else {
		logger.Info("Pulling artifact in", logger.Args("directory", o.destDir))
	}

	os, arch := runtime.GOOS, runtime.GOARCH
	if len(o.Artifact.Platforms) > 0 {
		os, arch = o.OSArch(0)
	}

	res, err := puller.Pull(ctx, ref, o.destDir, os, arch)
	if err != nil {
		return err
	}

	logger.Info("Artifact pulled", logger.Args("name", args[0], "type", res.Type, "digest", res.Digest))

	return nil
}
