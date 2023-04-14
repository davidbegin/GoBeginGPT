package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func GetGreatGreatGrandparentDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	parentDir := filepath.Dir(filename)
	grandparentDir := filepath.Dir(parentDir)
	greatGrandparentDir := filepath.Dir(grandparentDir)
	ggpd := filepath.Dir(greatGrandparentDir)
	return ggpd, nil
}

func GetGreatGrandparentDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	parentDir := filepath.Dir(filename)
	grandparentDir := filepath.Dir(parentDir)
	greatGrandparentDir := filepath.Dir(grandparentDir)
	return greatGrandparentDir, nil
}

func GetGrandparentDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	parentDir := filepath.Dir(filename)
	grandparentDir := filepath.Dir(parentDir)
	return grandparentDir, nil
}
