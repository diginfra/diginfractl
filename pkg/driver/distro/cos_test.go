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

package driverdistro

import (
	"testing"

	"github.com/diginfra/driverkit/pkg/kernelrelease"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

func TestDistroCosInit(t *testing.T) {
	type config struct {
		BuildID string `ini:"BUILD_ID"`
	}

	type testCase struct {
		cfg         interface{}
		buildID     string
		errExpected bool
	}
	testCases := []testCase{
		{
			cfg:     &config{BuildID: "17162.127.33"},
			buildID: "17162.127.33",
		},
		{
			cfg:         &struct{}{},
			errExpected: true,
		},
	}

	for _, tCase := range testCases {
		co := &cos{generic: &generic{}}
		cfg := ini.Empty()
		err := cfg.ReflectFrom(tCase.cfg)
		require.NoError(t, err)

		kr := kernelrelease.KernelRelease{}
		err = co.init(kr, "", cfg)
		if tCase.errExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tCase.buildID, co.buildID)
		}
	}
}
