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

package driverconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"

	"github.com/diginfra/diginfractl/internal/config"
	"github.com/diginfra/diginfractl/internal/utils"
	drivertype "github.com/diginfra/diginfractl/pkg/driver/type"
	"github.com/diginfra/diginfractl/pkg/options"
)

const (
	longConfig = `Configure a driver for future usages with other driver subcommands.
It will also update local Diginfra configuration or k8s configmap depending on the environment where it is running, to let Diginfra use chosen driver.
Only supports deployments of Diginfra that use a driver engine, ie: one between kmod, ebpf and modern-ebpf.
If engine.kind key is set to a non-driver driven engine, Diginfra configuration won't be touched.
`
)

type driverConfigOptions struct {
	*options.Common
	*options.Driver
	Update     bool
	Namespace  string
	KubeConfig string
}

type engineCfg struct {
	Kind string `yaml:"kind"`
}
type diginfraCfg struct {
	Engine engineCfg `yaml:"engine"`
}

// NewDriverConfigCmd configures a driver and stores it in config.
func NewDriverConfigCmd(ctx context.Context, opt *options.Common, driver *options.Driver) *cobra.Command {
	o := driverConfigOptions{
		Common: opt,
		Driver: driver,
	}

	cmd := &cobra.Command{
		Use:                   "config [flags]",
		DisableFlagsInUseLine: true,
		Short:                 "Configure a driver",
		Long:                  longConfig,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Override "namespace" flag with viper config if not set by user.
			f := cmd.Flags().Lookup("namespace")
			if f == nil {
				// should never happen
				return fmt.Errorf("unable to retrieve flag namespace")
			} else if !f.Changed && viper.IsSet(config.DriverNamespaceKey) {
				val := viper.Get(config.DriverNamespaceKey)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					return fmt.Errorf("unable to overwrite \"namespace\" flag: %w", err)
				}
			}

			// Override "update-diginfra" flag with viper config if not set by user.
			f = cmd.Flags().Lookup("update-diginfra")
			if f == nil {
				// should never happen
				return fmt.Errorf("unable to retrieve flag update-diginfra")
			} else if !f.Changed && viper.IsSet(config.DriverUpdateDiginfraKey) {
				val := viper.Get(config.DriverUpdateDiginfraKey)
				if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
					return fmt.Errorf("unable to overwrite \"update-diginfra\" flag: %w", err)
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.RunDriverConfig(ctx)
		},
	}

	cmd.Flags().BoolVar(&o.Update, "update-diginfra", true, "Whether to update Diginfra config/configmap.")
	cmd.Flags().StringVar(&o.Namespace, "namespace", "", "Kubernetes namespace.")
	cmd.Flags().StringVar(&o.KubeConfig, "kubeconfig", "", "Kubernetes config.")
	return cmd
}

// RunDriverConfig implements the driver configuration command.
func (o *driverConfigOptions) RunDriverConfig(ctx context.Context) error {
	o.Printer.Logger.Info("Running diginfractl driver config", o.Printer.Logger.Args(
		"name", o.Driver.Name,
		"version", o.Driver.Version,
		"type", o.Driver.Type.String(),
		"host-root", o.Driver.HostRoot,
		"repos", strings.Join(o.Driver.Repos, ",")))

	if o.Update {
		if err := o.commit(ctx, o.Driver.Type); err != nil {
			return err
		}
	}
	o.Printer.Logger.Info("Storing diginfractl driver config")
	return config.StoreDriver(o.Driver.ToDriverConfig(), o.ConfigFile)
}

func checkDiginfraRunsWithDrivers(engineKind string) error {
	// Modify the data in the ConfigMap/Diginfra config file ONLY if engine.kind is set to a known driver type.
	// This ensures that we modify the config only for Diginfras running with drivers, and not plugins/gvisor.
	// Scenario: user has multiple Diginfra pods deployed in its cluster, one running with driver,
	// other running with plugins. We must only touch the one running with driver.
	if _, err := drivertype.Parse(engineKind); err != nil {
		return fmt.Errorf("engine.kind is not driver driven: %s", engineKind)
	}
	return nil
}

func (o *driverConfigOptions) replaceDriverTypeInDiginfraConfig(driverType drivertype.DriverType) error {
	diginfraCfgFile := filepath.Clean(filepath.Join(string(os.PathSeparator), "etc", "diginfra", "diginfra.yaml"))
	yamlFile, err := os.ReadFile(filepath.Clean(diginfraCfgFile))
	if err != nil {
		return err
	}
	cfg := diginfraCfg{}
	if err = yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return err
	}
	if err = checkDiginfraRunsWithDrivers(cfg.Engine.Kind); err != nil {
		o.Printer.Logger.Warn("Avoid updating",
			o.Printer.Logger.Args("config", diginfraCfgFile, "reason", err))
		return nil
	}
	const configKindKey = "kind: "
	return utils.ReplaceTextInFile(diginfraCfgFile, configKindKey+cfg.Engine.Kind, configKindKey+driverType.String(), 1)
}

func (o *driverConfigOptions) replaceDriverTypeInK8SConfigMap(ctx context.Context, driverType drivertype.DriverType) error {
	var (
		err error
		cfg *rest.Config
	)

	if o.KubeConfig != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", o.KubeConfig)
	} else {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		return err
	}

	cl, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	configMapList, err := cl.CoreV1().ConfigMaps(o.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/instance=diginfra",
	})
	if err != nil {
		return err
	}
	if configMapList.Size() == 0 {
		return errors.New(`no configmaps matching "app.kubernetes.io/instance=diginfra" label were found`)
	}

	updated := false
	for i := 0; i < len(configMapList.Items); i++ {
		configMap := &configMapList.Items[i]
		// Check that this is a Diginfra config map
		diginfraYaml, present := configMap.Data["diginfra.yaml"]
		if !present {
			o.Printer.Logger.Debug("Skip non Diginfra-related config map",
				o.Printer.Logger.Args("configMap", configMap.Name))
			continue
		}

		// Check that Diginfra is configured to run with a driver
		var diginfraConfig diginfraCfg
		err = yaml.Unmarshal([]byte(diginfraYaml), &diginfraConfig)
		if err != nil {
			o.Printer.Logger.Warn("Failed to unmarshal diginfra.yaml to diginfraCfg struct",
				o.Printer.Logger.Args("configMap", configMap.Name, "err", err))
			continue
		}
		if err = checkDiginfraRunsWithDrivers(diginfraConfig.Engine.Kind); err != nil {
			o.Printer.Logger.Warn("Avoid updating",
				o.Printer.Logger.Args("configMap", configMap.Name, "reason", err))
			continue
		}

		// Update the configMap.
		// Multiple steps:
		// * unmarshal the `diginfra.yaml` as map[string]interface
		// * update `engine.kind` value
		// * save back the marshaled map to the `diginfra.yaml` configmap
		// * update the configmap
		var diginfraCfgData map[string]interface{}
		err = yaml.Unmarshal([]byte(diginfraYaml), &diginfraCfgData)
		if err != nil {
			o.Printer.Logger.Warn("Failed to unmarshal diginfra.yaml to map[string]interface{}",
				o.Printer.Logger.Args("configMap", configMap.Name, "err", err))
			continue
		}
		diginfraCfgEngine, ok := diginfraCfgData["engine"].(map[string]interface{})
		if !ok {
			o.Printer.Logger.Warn("Error fetching engine config",
				o.Printer.Logger.Args("configMap", configMap.Name))
			continue
		}
		diginfraCfgEngine["kind"] = driverType.String()
		diginfraCfgData["engine"] = diginfraCfgEngine
		diginfraCfgBytes, err := yaml.Marshal(diginfraCfgData)
		if err != nil {
			o.Printer.Logger.Warn("Error generating update data",
				o.Printer.Logger.Args("configMap", configMap.Name, "err", err))
			continue
		}
		configMap.Data["diginfra.yaml"] = string(diginfraCfgBytes)
		attempt := 0
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			o.Printer.Logger.Debug("Updating",
				o.Printer.Logger.Args("configMap", configMap.Name, "attempt", attempt))
			_, err := cl.CoreV1().ConfigMaps(configMap.Namespace).Update(
				ctx, configMap, metav1.UpdateOptions{})
			attempt++
			return err
		})
		if err != nil {
			return err
		}
		updated = true
	}

	if !updated {
		return errors.New("could not update any configmap")
	}
	return nil
}

// commit saves the updated driver type to Diginfra config,
// either to the local diginfra.yaml or updating the deployment configmap.
func (o *driverConfigOptions) commit(ctx context.Context, driverType drivertype.DriverType) error {
	if o.Namespace != "" {
		// Ok we are on k8s
		o.Printer.Logger.Info("Committing driver config to k8s configmap",
			o.Printer.Logger.Args("namespace", o.Namespace))
		return o.replaceDriverTypeInK8SConfigMap(ctx, driverType)
	}
	o.Printer.Logger.Info("Committing driver config to local Diginfra config")
	return o.replaceDriverTypeInDiginfraConfig(driverType)
}
