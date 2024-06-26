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

package follow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/diginfra/diginfractl/cmd/artifact/install"
	"github.com/diginfra/diginfractl/internal/config"
	"github.com/diginfra/diginfractl/internal/follower"
	"github.com/diginfra/diginfractl/pkg/index/index"
	"github.com/diginfra/diginfractl/pkg/oci"
	"github.com/diginfra/diginfractl/pkg/options"
	"github.com/diginfra/diginfractl/pkg/output"
)

const (
	timeout = time.Second * 5

	longFollow = `This command allows you to keep up-to-date one or more given artifacts.
It checks for updates on a periodic basis and then downloads and installs the latest version, 
as specified by the passed tags. 

Artifact references and flags are passed as arguments through:
- command line options
- environment variables
- configuration file
The arguments passed through these different modalities are prioritized in the following order:
command line options, environment variables, and finally the configuration file. This means that
if an argument is passed through multiple modalities, the value set in the command line options
will take precedence over the value set in environment variables, which will in turn take precedence
over the value set in the configuration file.
Please note that when passing multiple artifact references via an environment variable, they must be
separated by a semicolon ';' and the environment variable used for references is called
DIGINFRACT_ARTIFACT_FOLLOW_REFS. Other arguments, if passed through environment variables, should start
with "DIGINFRACTL_" and be followed by the hierarchical keys used in the configuration file separated by
an underscore "_".

A reference is either a simple name or a fully qualified reference ("<registry>/<repository>"), 
optionally followed by ":<tag>" (":latest" is assumed by default when no tag is given).

When providing just the name of the artifact, the command will search for the artifacts in 
the configured index files, and if found, it will use the registry and repository specified 
in the indexes.

Example - Install and follow "latest" tag of "k8saudit-rules" artifact by relying on index metadata:
	diginfractl artifact follow k8saudit-rules

Example - Install and follow all updates from "k8saudit-rules" 0.5.x release series:
	diginfractl artifact follow k8saudit-rules:0.5

Example - Install and follow "cloudtrail" plugins using a fully qualified reference:
	diginfractl artifact follow ghcr.io/diginfra/plugins/ruleset/k8saudit:latest
`
)

type artifactFollowOptions struct {
	*options.Common
	*options.Registry
	*options.Directory
	tmpDir           string
	every            time.Duration
	cron             string
	diginfraVersions string
	versions         config.DiginfraVersions
	timeout          time.Duration
	closeChan        chan bool
	allowedTypes     oci.ArtifactTypeSlice
	noVerify         bool
}

// NewArtifactFollowCmd returns the artifact follow command.
//
//nolint:gocyclo // unknown reason for cyclomatic complexity
func NewArtifactFollowCmd(ctx context.Context, opt *options.Common) *cobra.Command {
	o := artifactFollowOptions{
		Common:    opt,
		Registry:  &options.Registry{},
		Directory: &options.Directory{},
		closeChan: make(chan bool),
		versions:  config.DiginfraVersions{},
	}

	cmd := &cobra.Command{
		Use:   "follow [ref1 [ref2 ...]] [flags]",
		Short: "Install a list of artifacts and continuously checks if there are updates",
		Long:  longFollow,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Override "every" flag with viper config if not set by user.
			f := cmd.Flags().Lookup("every")
			if f == nil {
				// should never happen
				return fmt.Errorf("unable to retrieve flag every")
			} else if !f.Changed && viper.IsSet(config.ArtifactFollowEveryKey) {
				val := viper.Get(config.ArtifactFollowEveryKey)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					return fmt.Errorf("unable to overwrite \"every\" flag: %w", err)
				}
			}

			// Override "cron" flag with viper config if not set by user.
			f = cmd.Flags().Lookup("cron")
			if f == nil {
				// should never happen
				return fmt.Errorf("unable to retrieve flag cron")
			} else if !f.Changed && viper.IsSet(config.ArtifactFollowCronKey) {
				val := viper.Get(config.ArtifactFollowCronKey)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					return fmt.Errorf("unable to overwrite \"cron\" flag: %w", err)
				}
			}

			// Override "diginfra-versions" flag with viper config if not set by user.
			f = cmd.Flags().Lookup("diginfra-versions")
			if f == nil {
				// should never happen
				return fmt.Errorf("unable to retrieve flag diginfra-versions")
			} else if !f.Changed && viper.IsSet(config.ArtifactFollowDiginfraVersionsKey) {
				val := viper.Get(config.ArtifactFollowDiginfraVersionsKey)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					return fmt.Errorf("unable to overwrite \"diginfra-versions\" flag: %w", err)
				}
			}

			// Override "rulesfiles-dir" flag with viper config if not set by user.
			f = cmd.Flags().Lookup(options.FlagRulesFilesDir)
			if f == nil {
				// should never happen
				return fmt.Errorf("unable to retrieve flag %q", options.FlagRulesFilesDir)
			} else if !f.Changed && viper.IsSet(config.ArtifactFollowRulesfilesDirKey) {
				val := viper.Get(config.ArtifactFollowRulesfilesDirKey)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					return fmt.Errorf("unable to overwrite %q flag: %w", options.FlagRulesFilesDir, err)
				}
			}

			// Override "plugins-dir" flag with viper config if not set by user.
			f = cmd.Flags().Lookup(options.FlagPluginsFilesDir)
			if f == nil {
				// should never happen
				return fmt.Errorf("unable to retrieve flag %q", options.FlagPluginsFilesDir)
			} else if !f.Changed && viper.IsSet(config.ArtifactFollowPluginsDirKey) {
				val := viper.Get(config.ArtifactFollowPluginsDirKey)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					return fmt.Errorf("unable to overwrite %q flag: %w", options.FlagPluginsFilesDir, err)
				}
			}

			// Override "assets-dir" flag with viper config if not set by user.
			f = cmd.Flags().Lookup(options.FlagAssetsFilesDir)
			if f == nil {
				// should never happen
				return fmt.Errorf("unable to retrieve flag %q", options.FlagAssetsFilesDir)
			} else if !f.Changed && viper.IsSet(config.ArtifactFollowAssetsDirKey) {
				val := viper.Get(config.ArtifactFollowAssetsDirKey)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					return fmt.Errorf("unable to overwrite %q flag: %w", options.FlagAssetsFilesDir, err)
				}
			}

			// Override "tmp-dir" flag with viper config if not set by user.
			f = cmd.Flags().Lookup("tmp-dir")
			if f == nil {
				// should never happen
				return fmt.Errorf("unable to retrieve flag tmp-dir")
			} else if !f.Changed && viper.IsSet(config.ArtifactFollowTmpDirKey) {
				val := viper.Get(config.ArtifactFollowTmpDirKey)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					return fmt.Errorf("unable to overwrite \"tmp-dir\" flag: %w", err)
				}
			}

			// Override "allowed-types" flag with viper config if not set by user.
			f = cmd.Flags().Lookup(install.FlagAllowedTypes)
			if f == nil {
				// should never happen
				return fmt.Errorf("unable to retrieve flag %s", install.FlagAllowedTypes)
			} else if !f.Changed && viper.IsSet(config.ArtifactAllowedTypesKey) {
				val, err := config.ArtifactAllowedTypes()
				if err != nil {
					return err
				}
				if err := cmd.Flags().Set(f.Name, val.String()); err != nil {
					return fmt.Errorf("unable to overwrite %q flag: %w", install.FlagAllowedTypes, err)
				}
			}

			// Override "no-verify" flag with viper config if not set by user.
			f = cmd.Flags().Lookup(install.FlagNoVerify)
			if f == nil {
				// should never happen
				return fmt.Errorf("unable to retrieve flag %s", install.FlagNoVerify)
			} else if !f.Changed && viper.IsSet(config.ArtifactNoVerifyKey) {
				val := viper.Get(config.ArtifactNoVerifyKey)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					return fmt.Errorf("unable to overwrite %q flag: %w", install.FlagNoVerify, err)
				}
			}

			// Get Diginfra versions via HTTP endpoint
			if err := o.retrieveDiginfraVersions(ctx); err != nil {
				return fmt.Errorf("unable to retrieve Diginfra versions, please check if it is running "+
					"and correctly exposing the version endpoint: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.RunArtifactFollow(ctx, args)
		},
	}

	o.Registry.AddFlags(cmd)
	o.Directory.AddFlags(cmd)
	cmd.Flags().DurationVarP(&o.every, "every", "e", config.FollowResync, "Time interval how often it checks for a new version of the "+
		"artifact. Cannot be used together with 'cron' option.")
	cmd.Flags().StringVar(&o.cron, "cron", "", "Cron-like string to specify interval how often it checks for a new version of the artifact."+
		" Cannot be used together with 'every' option.")
	cmd.Flags().StringVar(&o.tmpDir, "tmp-dir", "", "Directory where to save temporary files")
	cmd.Flags().StringVar(&o.diginfraVersions, "diginfra-versions", "http://localhost:8765/versions",
		"Where to retrieve versions, it can be either an URL or a path to a file")
	cmd.Flags().DurationVar(&o.timeout, "timeout", defaultBackoffConfig.MaxDelay,
		"Timeout for initial connection to the Diginfra versions endpoint")
	cmd.Flags().Var(&o.allowedTypes, install.FlagAllowedTypes,
		fmt.Sprintf(`list of artifact types that can be followed. If not specified or configured, all types are allowed.
It accepts comma separated values or it can be repeated multiple times.
Examples: 
	--%s="rulesfile,plugin"
	--%s=rulesfile --%s=plugin`, install.FlagAllowedTypes, install.FlagAllowedTypes, install.FlagAllowedTypes))
	cmd.Flags().BoolVar(&o.noVerify, install.FlagNoVerify, false,
		"whether this command should skip signature verification")
	cmd.MarkFlagsMutuallyExclusive("cron", "every")

	return cmd
}

// RunArtifactFollow executes the business logic for the artifact follow command.
func (o *artifactFollowOptions) RunArtifactFollow(ctx context.Context, args []string) error {
	logger := o.Printer.Logger
	// Retrieve configuration for follower
	configuredFollower, err := config.Follower()
	if err != nil {
		o.Printer.CheckErr(fmt.Errorf("unable to retrieved the configured follower: %w", err))
	}

	// Set args as configured if no arg was passed
	if len(args) == 0 {
		if len(configuredFollower.Artifacts) == 0 {
			return fmt.Errorf("no artifacts to follow, please configure artifacts or pass them as arguments to this command")
		}
		args = configuredFollower.Artifacts
	}

	var sched cron.Schedule
	if o.cron != "" {
		sched, err = cron.ParseStandard(o.cron)
		if err != nil {
			return fmt.Errorf("unable to parse cron '%s': %w", o.cron, err)
		}
	} else {
		sched = scheduledDuration{o.every}
	}

	var wg sync.WaitGroup
	// For each artifact create a follower.
	var followers = make(map[string]*follower.Follower, 0)
	for _, a := range args {
		if o.cron != "" {
			logger.Info("Creating follower", logger.Args("artifact", a, "cron", o.cron))
		} else {
			logger.Info("Creating follower", logger.Args("artifact", a, "check every", o.every.String()))
		}
		ref, err := o.IndexCache.ResolveReference(a)
		if err != nil {
			return fmt.Errorf("unable to parse artifact reference for %q: %w", a, err)
		}

		var sig *index.Signature
		if !o.noVerify {
			sig = o.IndexCache.SignatureForIndexRef(a)
		}

		cfg := &follower.Config{
			WaitGroup:         &wg,
			Resync:            sched,
			RulesfilesDir:     o.RulesfilesDir,
			PluginsDir:        o.PluginsDir,
			AssetsDir:         o.AssetsDir,
			ArtifactReference: ref,
			PlainHTTP:         o.PlainHTTP,
			CloseChan:         o.closeChan,
			TmpDir:            o.tmpDir,
			DiginfraVersions:  o.versions,
			AllowedTypes:      o.allowedTypes,
			Signature:         sig,
		}
		fol, err := follower.New(ref, o.Printer, cfg)
		if err != nil {
			return fmt.Errorf("unable to create the follower for ref %q: %w", ref, err)
		}
		wg.Add(1)
		followers[ref] = fol
	}

	for k, f := range followers {
		logger.Info("Starting follower", logger.Args("artifact", k))
		go f.Follow(ctx)
	}

	// Wait until we receive a signal to be terminated
	<-ctx.Done()

	// We are done, shutdown the followers.
	logger.Info("Closing followers...")
	close(o.closeChan)

	// Wait for the followers to shutdown or that the timer expires.
	doneChan := make(chan bool)

	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		logger.Info("Followers correctly stopped.")
	case <-time.After(timeout):
		logger.Info("Timed out waiting for followers to exit")
	}

	return nil
}

func (o *artifactFollowOptions) retrieveDiginfraVersions(ctx context.Context) error {
	_, err := url.ParseRequestURI(o.diginfraVersions)
	if err != nil {
		return fmt.Errorf("unable to parse URI: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.diginfraVersions, http.NoBody)
	if err != nil {
		return fmt.Errorf("cannot fetch Diginfra version: %w", err)
	}

	backoffConfig := defaultBackoffConfig
	backoffConfig.MaxDelay = o.timeout

	client := &http.Client{
		Transport: &backoffTransport{
			Base:    http.DefaultTransport,
			Printer: o.Printer,
			Config:  backoffConfig,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to get versions from URL %q: %w", o.diginfraVersions, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %w", err)
	}

	var dataUnmarshalled map[string]interface{}

	err = json.Unmarshal(data, &dataUnmarshalled)
	if err != nil {
		return fmt.Errorf("error unmarshalling: %w", err)
	}

	for key, value := range dataUnmarshalled {
		// todo(alacuku): how to handle types other than strings? Silently ignoring for now...
		if strValue, ok := value.(string); ok {
			o.versions[key] = strValue
		}
	}

	return nil
}

// Config defines the configuration options for backoff.
type backoffConfig struct {
	// BaseDelay is the amount of time to backoff after the first failure.
	BaseDelay time.Duration
	// Multiplier is the factor with which to multiply backoffs after a
	// failed retry. Should ideally be greater than 1.
	Multiplier float64
	// Jitter is the factor with which backoffs are randomized.
	// todo: not yet implemented
	// Jitter float64
	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration
}

var defaultBackoffConfig = backoffConfig{
	BaseDelay:  1.0 * time.Second,
	Multiplier: 1.6,
	// Jitter:     0.2, todo: not yet implemented
	MaxDelay: 120 * time.Second,
}

type backoffTransport struct {
	Base      http.RoundTripper
	Printer   *output.Printer
	Config    backoffConfig
	attempts  int
	startTime time.Time
}

func (bt *backoffTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var err error
	var resp *http.Response
	logger := bt.Printer.Logger
	bt.startTime = time.Now()
	bt.attempts = 0

	logger.Debug(fmt.Sprintf("Retrieving versions from Diginfra (timeout %s) ...", bt.Config.MaxDelay))

	for {
		resp, err = bt.Base.RoundTrip(req)
		if err != nil {
			if req.Context().Err() != nil {
				return nil, req.Context().Err()
			}
			sleep := bt.Config.backoff(bt.attempts)

			wakeTime := time.Now().Add(sleep)
			if wakeTime.Sub(bt.startTime) > bt.Config.MaxDelay {
				return resp, fmt.Errorf("timeout occurred while retrieving versions from Diginfra")
			}

			logger.Debug(fmt.Sprintf("error: %s. Trying again in %s", err.Error(), sleep.String()))
			time.Sleep(sleep)
		} else {
			logger.Debug("Successfully retrieved versions from Diginfra")
			return resp, err
		}

		bt.attempts++
	}
}

// Backoff returns the amount of time to wait before the next retry given the
// number of retries.
func (bc backoffConfig) backoff(retries int) time.Duration {
	if retries == 0 {
		return bc.BaseDelay
	}
	backoff, max := float64(bc.BaseDelay), float64(bc.MaxDelay)
	for backoff < max && retries > 0 {
		backoff *= bc.Multiplier
		retries--
	}
	if backoff > max {
		backoff = max
	}
	// Randomize backoff delays so that if a cluster of requests start at
	// the same time, they won't operate in lockstep.
	// todo: implement jitter
	// backoff *= 1 + bc.Jitter*(math.Float64()*2-1)
	if backoff < 0 {
		return 0
	}

	return time.Duration(backoff)
}

type scheduledDuration struct {
	time.Duration
}

func (sd scheduledDuration) Next(tm time.Time) time.Time {
	return tm.Add(sd.Duration)
}
