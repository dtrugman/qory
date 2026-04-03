package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dtrugman/qory/lib/util"
)

func modelsHaveProviders(models []string) bool {
	for _, model := range models {
		if !strings.Contains(model, "/") {
			return false
		}
	}
	return true
}

func promptModelsWithProviders(models []string) (string, error) {
	providers := make(map[string][]string)
	for _, model := range models {
		parts := strings.SplitN(model, "/", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("model missing provider: %s", model)
		}
		providers[parts[0]] = append(providers[parts[0]], parts[1])
	}

	providerNames := util.MapKeys(providers)
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

func promptModelsWithoutProviders(models []string) (string, error) {
	sort.Strings(models)
	return promptFromList(models)
}

func promptModel(models []string) (string, error) {
	if modelsHaveProviders(models) {
		return promptModelsWithProviders(models)
	} else {
		return promptModelsWithoutProviders(models)
	}
}
