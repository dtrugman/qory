package main

import (
	"strings"

	"github.com/dtrugman/qory/lib/config"
)

type paramSetterBaseURL struct {
	paramSetter
}

func (s *paramSetterBaseURL) AdjustValue(value *string) error {
	if !strings.HasSuffix(*value, "/") {
		*value += "/"
	}

	return nil
}

func NewParamBaseURL(conf config.Config, key string) Param {
	return &paramBase{
		conf:   conf,
		key:    key,
		setter: &paramSetterBaseURL{},
	}
}
