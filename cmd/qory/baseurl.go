package main

import (
	"fmt"
	"strings"

	"github.com/dtrugman/qory/lib/config"
)

type paramSetterBaseURL struct {
	paramSetter
}

func (s *paramSetterBaseURL) validateBaseURL(value string) error {
	if !strings.HasSuffix(value, "/") {
		return fmt.Errorf("must end with a '/'")
	} else {
		return nil
	}
}

func NewParamBaseURL(conf config.Config, key string) Param {
	return &paramBase{
		conf:   conf,
		key:    key,
		setter: &paramSetterBaseURL{},
	}
}
