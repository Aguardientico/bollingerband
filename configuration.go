package main

import (
	"encoding/json"
	"os"
)

type ConfigurationInfo struct {
	Symbols []string
	Periods int
	Factor  float64
}

// Read config.json file to capture configuration options
func Configuration() ConfigurationInfo {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(file)
	configuration := &ConfigurationInfo{}
	err = decoder.Decode(&configuration)
	if err != nil {
		panic(err)
	}
	return *configuration
}
