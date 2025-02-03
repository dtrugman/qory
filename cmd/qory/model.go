package main

import (
	"fmt"
	"iter"
	"maps"
	"sort"
	"strings"

	"github.com/dtrugman/qory/lib/config"
	"github.com/dtrugman/qory/lib/model"
)

type paramSetterModel struct {
	paramSetter

	client model.Client
}

func NewParamModel(conf config.Config, key string, client model.Client) Param {
	return &paramBase{
		conf:   conf,
		key:    key,
		setter: &paramSetterModel{client: client},
	}
}

func seqToSlice[T any](seq iter.Seq[T]) []T {
	var result []T
	for value := range seq {
		result = append(result, value)
	}
	return result
}

func allHaveProviders(models []string) bool {
	for _, model := range models {
		if !strings.Contains(model, "/") {
			return false
		}
	}

	return true
}

func (p *paramSetterModel) promptModelsWithProviders(models []string) (string, error) {
	providers := make(map[string][]string, 0)
	for _, model := range models {
		parts := strings.SplitN(model, "/", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("model missing provider: %s", model)
		}

		providerName := parts[0]
		modelName := parts[1]

		providers[providerName] = append(providers[providerName], modelName)
	}

	providerIter := maps.Keys(providers)
	providerNames := seqToSlice(providerIter)
	sort.Strings(providerNames)

	providerSelected, err := promptFromList(providerNames)
	if err != nil {
		return "", err
	}

	modelNames := providers[providerSelected]
	sort.Strings(modelNames)

	modelSelected, err := promptFromList(modelNames)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", providerSelected, modelSelected), nil
}

func (p *paramSetterModel) promptModelsWithoutProviders(models []string) (string, error) {
	sort.Strings(models)
	return promptFromList(models)
}

func (p *paramSetterModel) PromptValue() (string, error) {
	models, err := p.client.AvailableModels()
	if err != nil {
		return "", err
	}

	if allHaveProviders(models) {
		return p.promptModelsWithProviders(models)
	} else {
		return p.promptModelsWithoutProviders(models)
	}
}
