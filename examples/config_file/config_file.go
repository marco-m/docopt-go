package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/marco-m/docopt-go"
)

func loadJSONConfig() map[string]any {
	var result map[string]any
	jsonData := []byte(`{"--force": true, "--timeout": "10", "--baud": "9600"}`)
	json.Unmarshal(jsonData, &result)
	return result
}

func loadIniConfig() map[string]any {
	iniData := `
[default-arguments]
--force
--baud=19200
<host>=localhost`
	// trivial ini parser
	// default value for an item is bool: true (for --force)
	// otherwise the value is a string
	iniParsed := make(map[string]map[string]any)
	var section string
	for _, line := range strings.Split(iniData, "\n") {
		if strings.HasPrefix(line, "[") {
			section = line
			iniParsed[section] = make(map[string]any)
		} else if section != "" {
			kv := strings.SplitN(line, "=", 2)
			if len(kv) == 1 {
				iniParsed[section][kv[0]] = true
			} else if len(kv) == 2 {
				iniParsed[section][kv[0]] = kv[1]
			}
		}
	}
	return iniParsed["[default-arguments]"]
}

// merge combines two maps.
// truthiness takes priority over falsiness
// mapA takes priority over mapB
func merge(mapA, mapB map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range mapA {
		result[k] = v
	}
	for k, v := range mapB {
		if _, ok := result[k]; !ok || result[k] == nil || result[k] == false {
			result[k] = v
		}
	}
	return result
}

func main() {
	usage := `Usage:
  config_file tcp [<host>] [--force] [--timeout=<seconds>]
  config_file serial <port> [--baud=<rate>] [--timeout=<seconds>]
  config_file -h | --help | --version`

	jsonConfig := loadJSONConfig()
	iniConfig := loadIniConfig()
	arguments, _ := docopt.Parse(usage, nil, "0.1.1rc")

	// Arguments take priority over INI, INI takes priority over JSON
	result := merge(arguments, merge(iniConfig, jsonConfig))

	fmt.Println("JSON config: ", jsonConfig)
	fmt.Println("INI config: ", iniConfig)
	fmt.Println("Result: ", result)
}
