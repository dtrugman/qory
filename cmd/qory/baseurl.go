package main

import (
	"strings"
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

func NewParamBaseURL(conf Config, key string) Param {
	return &paramBase{
		conf:   conf,
		key:    key,
		setter: &paramSetterBaseURL{},
	}
}
