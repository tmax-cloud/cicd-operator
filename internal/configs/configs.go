/*
 Copyright 2021 The CI/CD Operator Authors

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

package configs

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

// Initiate-related channel
var (
	ControllerInitiated = false
	ControllerInitCh    = make(chan struct{}, 1)

	BlockerInitiated = false
	BlockerInitCh    = make(chan struct{}, 1)
)

type cfgType int

const (
	cfgTypeString cfgType = iota
	cfgTypeInt
	cfgTypeBool
)

type operatorConfig struct {
	Type cfgType

	StringVal     *string
	StringDefault string

	IntVal     *int
	IntDefault int

	BoolVal     *bool
	BoolDefault bool
}

// Handler is a config map handler function
type Handler func(cm *corev1.ConfigMap) error

func getVars(data map[string]string, vars map[string]operatorConfig) {
	for key, c := range vars {
		v, exist := data[key]
		switch c.Type {
		case cfgTypeString:
			if c.StringVal == nil {
				continue
			}
			if exist {
				*c.StringVal = v
			} else {
				*c.StringVal = c.StringDefault
			}
		case cfgTypeInt:
			if c.IntVal == nil {
				continue
			}
			if exist {
				i, err := strconv.Atoi(v)
				if err != nil {
					continue
				}
				*c.IntVal = i
			} else {
				*c.IntVal = c.IntDefault
			}
		case cfgTypeBool:
			if c.BoolVal == nil {
				continue
			}
			if exist {
				b, err := strconv.ParseBool(v)
				if err != nil {
					continue
				}
				*c.BoolVal = b
			} else {
				*c.BoolVal = c.BoolDefault
			}
		}
	}
}
