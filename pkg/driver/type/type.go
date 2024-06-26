// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2024 The Diginfra Authors
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

// Package drivertype implements all the driver type specific logic.
package drivertype

import (
	"fmt"

	"github.com/diginfra/driverkit/cmd"
	"github.com/diginfra/driverkit/pkg/kernelrelease"

	"github.com/diginfra/diginfractl/pkg/output"
)

// KernelDirEnv is the env variable set to kernel headers extraction paths.
const KernelDirEnv = "KERNELDIR"

var driverTypes = map[string]DriverType{}

// DriverType is the interface that wraps driver types.
type DriverType interface {
	fmt.Stringer
	Cleanup(printer *output.Printer, driverName string) error
	Load(printer *output.Printer, src, driverName string, fallback bool) error
	Extension() string
	HasArtifacts() bool
	ToOutput(destPath string) cmd.OutputOptions
	Supported(kr kernelrelease.KernelRelease) bool
}

// GetTypes return the list of supported driver types.
func GetTypes() []string {
	driverTypesSlice := make([]string, 0)
	for key := range driverTypes {
		driverTypesSlice = append(driverTypesSlice, key)
	}
	return driverTypesSlice
}

// Parse parses a driver type string and returns the corresponding DriverType object or an error.
func Parse(driverType string) (DriverType, error) {
	if dType, ok := driverTypes[driverType]; ok {
		return dType, nil
	}
	return nil, fmt.Errorf("unsupported driver type specified: %s", driverType)
}
