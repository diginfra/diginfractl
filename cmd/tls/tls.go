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

package tls

import (
	"github.com/spf13/cobra"

	"github.com/diginfra/diginfractl/cmd/tls/install"
	commonoptions "github.com/diginfra/diginfractl/pkg/options"
)

// NewTLSCmd return the tls command.
func NewTLSCmd(opt *commonoptions.Common) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "tls",
		TraverseChildren:      true,
		DisableFlagsInUseLine: true,
		Short:                 "Generate and install TLS material for Diginfra",
		Long:                  `Generate and install TLS material for Diginfra`,
	}

	cmd.AddCommand(install.NewTLSInstallCmd(opt))

	return cmd
}
