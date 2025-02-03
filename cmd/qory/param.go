package main

import (
	"fmt"

	"github.com/dtrugman/qory/lib/config"
)

type paramSetter struct {
}

type ParamSetter interface {
	PromptValue() (string, error)
	ValidateValue(value string) error
	AdjustValue(value *string) error
}

func (s *paramSetter) PromptValue() (string, error) {
	return promptUserInput()
}

func (s *paramSetter) ValidateValue(value string) error {
	return nil
}

func (s *paramSetter) AdjustValue(value *string) error {
	return nil
}

type paramBase struct {
	conf   config.Config
	key    string
	setter ParamSetter
}

type Param interface {
	Get() error
	Unset() error
	Set(value *string) error
}

func NewParam(conf config.Config, key string) Param {
	return &paramBase{
		conf:   conf,
		key:    key,
		setter: &paramSetter{},
	}
}

func (p *paramBase) Get() error {
	if value, err := p.conf.Get(p.key); err != nil {
		return err
	} else if value == nil {
		fmt.Printf("No value\n")
		return nil
	} else {
		fmt.Printf("%s\n", *value)
		return nil
	}
}

func (p *paramBase) Unset() error {
	if err := p.conf.Unset(p.key); err != nil {
		return err
	} else {
		fmt.Printf("DONE\n")
		return nil
	}
}

func (p *paramBase) Set(
	value *string,
) error {
	if value == nil {
		input, err := p.setter.PromptValue()
		if err != nil {
			return err
		}
		value = &input
	}

	if err := p.setter.ValidateValue(*value); err != nil {
		return err
	}

	if err := p.setter.AdjustValue(value); err != nil {
		return err
	}

	if err := p.conf.Set(p.key, *value); err != nil {
		return err
	}

	fmt.Printf("DONE\n")
	return nil
}
