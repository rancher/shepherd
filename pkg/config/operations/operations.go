package operations

import (
	"encoding/json"
	"errors"
	"fmt"

	"sigs.k8s.io/yaml"
)

// ReplaceValue recursively traverses keyPath, replacing the specified replaceVal in searchMap.
func ReplaceValue(keyPath []string, replaceVal any, searchMap map[string]any) (map[string]any, error) {
	if len(keyPath) <= 1 {
		searchMap[keyPath[0]] = replaceVal

		return searchMap, nil
	} else {
		var err error

		if _, ok := searchMap[keyPath[0]].(map[string]any); ok {
			searchMap[keyPath[0]], err = ReplaceValue(keyPath[1:], replaceVal, searchMap[keyPath[0]].(map[string]any))
			if err != nil {
				return nil, err
			}
		} else if _, ok := searchMap[keyPath[0]].([]any); ok {
			for i := range searchMap[keyPath[0]].([]any) {
				searchMap[keyPath[0]].([]any)[i], err = ReplaceValue(keyPath[1:], replaceVal, searchMap[keyPath[0]].([]any)[i].(map[string]any))
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return searchMap, nil
}

// GetValue recursively traverses keyPath, returning the value of the specified keyPath in searchMap.
func GetValue(keyPath []string, searchMap map[string]any) (any, error) {
	var err error
	var keypathvalues any
	if len(keyPath) == 1 {
		keypathvalues, ok := searchMap[keyPath[0]]
		if !ok {
			err = errors.New(fmt.Sprintf("expected key does not exist: %s", keyPath[0]))
		}
		return keypathvalues, err
	} else {
		if _, ok := searchMap[keyPath[0]].(map[string]any); ok {
			keypathvalues, err = GetValue(keyPath[1:], searchMap[keyPath[0]].(map[string]any))
			if err != nil {
				return nil, err
			}
		} else if _, ok := searchMap[keyPath[0]].([]any); ok {
			for i := range searchMap[keyPath[0]].([]any) {
				keypathvalues, err = GetValue(keyPath[1:], searchMap[keyPath[0]].([]any)[i].(map[string]any))
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return keypathvalues, err
}

// LoadObjectFromMap unmarshals a specific key's value into an object
func LoadObjectFromMap(key string, config map[string]any, object any) {
	keyConfig := config[key]
	scopedString, err := yaml.Marshal(keyConfig)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(scopedString, &object)
	if err != nil {
		panic(err)
	}
}

// DeepCopyMap creates a copy of the map that doesn't have any links to the original map
func DeepCopyMap(config map[string]any) (map[string]any, error) {
	marshaledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	unmarshaledConfig := make(map[string]any)
	err = json.Unmarshal(marshaledConfig, &unmarshaledConfig)
	if err != nil {
		return nil, err
	}

	return unmarshaledConfig, nil
}
