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

package oci

const (

	// DiginfraRulesfileConfigMediaType is the MediaType for rule's config layer.
	DiginfraRulesfileConfigMediaType = "application/vnd.cncf.diginfra.rulesfile.config.v1+json"

	// DiginfraRulesfileLayerMediaType is the MediaType for rules.
	DiginfraRulesfileLayerMediaType = "application/vnd.cncf.diginfra.rulesfile.layer.v1+tar.gz"

	// DiginfraPluginConfigMediaType is the MediaType for plugin's config layer.
	DiginfraPluginConfigMediaType = "application/vnd.cncf.diginfra.plugin.config.v1+json"

	// DiginfraPluginLayerMediaType is the MediaType for plugins.
	DiginfraPluginLayerMediaType = "application/vnd.cncf.diginfra.plugin.layer.v1+tar.gz"

	// DiginfraAssetConfigMediaType is the MediaType for asset's config layer.
	DiginfraAssetConfigMediaType = "application/vnd.cncf.diginfra.asset.config.v1+json"

	// DiginfraAssetLayerMediaType is the MediaType for assets.
	DiginfraAssetLayerMediaType = "application/vnd.cncf.diginfra.asset.layer.v1+tar.gz"

	// DefaultTag is the default tag reference to be used when none is provided.
	DefaultTag = "latest"
)
