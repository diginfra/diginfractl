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

package config_test

import (
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"github.com/diginfra/diginfractl/cmd"
)

var usage = `Usage:
  diginfractl artifact config [ref] [flags]

Flags:
  -h, --help         help for config
      --plain-http   allows interacting with remote registry via plain http requests

Global Flags:
      --config string       config file to be used for diginfractl (default "/etc/diginfractl/diginfractl.yaml")
      --log-format string   Set formatting for logs (color, text, json) (default "color")
      --log-level string    Set level for logs (info, warn, debug, trace) (default "info")
`

var help = `Get the config layer of an artifact

Usage:
  diginfractl artifact config [ref] [flags]

Flags:
  -h, --help              help for config
      --plain-http        allows interacting with remote registry via plain http requests
      --platform string   os and architecture of the artifact in OS/ARCH format (default "linux/amd64")

Global Flags:
      --config string       config file to be used for diginfractl (default "/etc/diginfractl/diginfractl.yaml")
      --log-format string   Set formatting for logs (color, text, json) (default "color")
      --log-level string    Set level for logs (info, warn, debug, trace) (default "info")
`

var _ = Describe("Config", func() {
	const (
		artifactCmd  = "artifact"
		configCmd    = "config"
		plaingHTTP   = "--plain-http"
		configFlag   = "--config"
		platformFlag = "--platform"
	)

	var (
		err       error
		args      []string
		configDir string
	)

	var assertFailedBehavior = func(usage, specificError string) {
		It("check that fails and the usage is not printed", func() {
			Expect(err).To(HaveOccurred())
			Expect(output).ShouldNot(gbytes.Say(regexp.QuoteMeta(usage)))
			Expect(output).Should(gbytes.Say(regexp.QuoteMeta(specificError)))
		})
	}

	JustBeforeEach(func() {
		configDir = GinkgoT().TempDir()
		rootCmd = cmd.New(ctx, opt)
		err = executeRoot(args)
	})

	JustAfterEach(func() {
		err = nil
		Expect(output.Clear()).ShouldNot(HaveOccurred())
		args = nil
	})

	Context("help message", func() {
		BeforeEach(func() {
			args = []string{artifactCmd, configCmd, "--help"}
		})

		It("should match the saved one", func() {
			Expect(err).ShouldNot(HaveOccurred())
			Expect(string(output.Contents())).Should(Equal(help))
		})
	})

	Context("wrong number of arguments", func() {
		When("number of arguments equal to 0", func() {
			BeforeEach(func() {
				args = []string{artifactCmd, configCmd}
			})

			assertFailedBehavior(usage, "ERROR accepts 1 arg(s), received 0 ")
		})

		When("number of arguments equal to 2", func() {
			BeforeEach(func() {
				args = []string{artifactCmd, configCmd, "arg1", "arg2", configFlag, configDir}
			})

			assertFailedBehavior(usage, "ERROR accepts 1 arg(s), received 2 ")
		})
	})

	Context("failure", func() {
		When("unreachable/non existing registry", func() {
			BeforeEach(func() {
				args = []string{artifactCmd, configCmd, "noregistry/noartifact", plaingHTTP, configFlag, configDir}
			})

			assertFailedBehavior(usage, "ERROR unable to get manifest: unable to fetch reference")
		})

		When("non existing repository", func() {
			BeforeEach(func() {
				args = []string{artifactCmd, configCmd, localRegistryHost + "/noartifact", plaingHTTP, configFlag, configDir}
			})

			assertFailedBehavior(usage, "noartifact:latest: not found")
		})

		When("non parsable reference", func() {
			BeforeEach(func() {
				args = []string{artifactCmd, configCmd, " ", plaingHTTP, configFlag, configDir}
			})

			assertFailedBehavior(usage, "ERROR cannot find   among the configured indexes, skipping ")
		})

		When("no manifest for given platform", func() {
			BeforeEach(func() {
				args = []string{artifactCmd, configCmd, pluginMultiPlatformRef, plaingHTTP, configFlag, configDir, platformFlag, "linux/wrong"}
			})
			assertFailedBehavior(usage, "ERROR unable to get manifest: unable to find a manifest matching the given platform: linux/wrong")
		})
	})

	Context("success", func() {
		When("empty config layer", func() {
			BeforeEach(func() {
				args = []string{artifactCmd, configCmd, artifactWithoutConfigRef, plaingHTTP, configFlag, configDir}
			})

			It("should success", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(output).Should(gbytes.Say(regexp.QuoteMeta("{}")))
			})
		})

		When("with valid config layer", func() {
			BeforeEach(func() {
				args = []string{artifactCmd, configCmd, rulesRef, plaingHTTP, configFlag, configDir}
			})

			It("should success", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(output).Should(gbytes.Say(regexp.QuoteMeta(`{"dependencies":[{"name":"dep1","version":"1.2.3"},{"name":"dep2","version":"2.3.1"}]}`)))
			})
		})

		When("no platform flag", func() {
			BeforeEach(func() {
				args = []string{artifactCmd, configCmd, pluginMultiPlatformRef, plaingHTTP, configFlag, configDir}
			})

			It("should success getting the platform where tests are running", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(output).Should(gbytes.Say(regexp.QuoteMeta(
					`{"dependencies":[{"name":"my-dep","version":"1.2.3","alternatives":[{"name":"my-alt-dep","version":"`)))
			})
		})

		When("with valid platform", func() {
			BeforeEach(func() {
				args = []string{artifactCmd, configCmd, pluginMultiPlatformRef, plaingHTTP, configFlag, configDir, platformFlag, testPluginPlatform3}
			})

			It("should success", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(output).Should(gbytes.Say(regexp.QuoteMeta(
					`{"dependencies":[{"name":"my-dep","version":"1.2.3","alternatives":[{"name":"my-alt-dep","version":"`)))
			})
		})

		When("with non existing platform for artifacts without platforms", func() {
			BeforeEach(func() {
				args = []string{artifactCmd, configCmd, rulesRef, plaingHTTP, configFlag, configDir, platformFlag, testPluginPlatform3}
			})

			It("should success", func() {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(output).Should(gbytes.Say(regexp.QuoteMeta(`{"dependencies":[{"name":"dep1","version":"1.2.3"},{"name":"dep2","version":"2.3.1"}]}`)))
			})
		})

	})
})
