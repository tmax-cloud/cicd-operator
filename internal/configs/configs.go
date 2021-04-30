package configs

import (
	corev1 "k8s.io/api/core/v1"
	"strconv"
)

// Initiate-related channel
var (
	Initiated = false
	InitCh    = make(chan struct{}, 1)
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
		v := data[key]
		switch c.Type {
		case cfgTypeString:
			if c.StringVal == nil {
				continue
			}
			if len(v) > 0 {
				*c.StringVal = v
			} else if len(c.StringDefault) > 0 {
				*c.StringVal = c.StringDefault
			}
		case cfgTypeInt:
			if c.IntVal == nil {
				continue
			}
			if len(v) > 0 {
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
			if len(v) > 0 {
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
