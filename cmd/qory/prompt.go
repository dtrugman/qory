package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func promptUserInput() (string, error) {
	fmt.Print("Enter value: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

func promptFromList(list []string) (string, error) {
	for i, value := range list {
		fmt.Printf("%d. %s\n", i+1, value)
	}

	fmt.Print("Choose option: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSuffix(input, "\n")

	index, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid number")
	}

	index = index - 1
	if index < 0 || index >= len(list) {
		return "", fmt.Errorf("bad selection")
	}

	return list[index], nil
}
