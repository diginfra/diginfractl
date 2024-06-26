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

package install

const (
	// FlagAllowedTypes is the name of the flag to specify allowed artifact types.
	FlagAllowedTypes = "allowed-types"

	// FlagPlatform is the name of the flag to override the platform.
	FlagPlatform = "platform"

	// FlagResolveDeps is the name of the flag to enable artifact dependencies resolution.
	FlagResolveDeps = "resolve-deps"

	// FlagNoVerify is the name of the flag to disable signature verification.
	FlagNoVerify = "no-verify"
)
