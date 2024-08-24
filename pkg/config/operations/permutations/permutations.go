package permutations

import (
	"errors"

	mapOperations "github.com/rancher/shepherd/pkg/config/operations"
)

// Relationship structs are used to create an association between a parent value and either a set of permutations or the value of a different key
type Relationship struct {
	ParentValue       any           `json:"parentValue" yaml:"parentValue"`
	ChildKeyPath      []string      `json:"childKeyPath" yaml:"childkeyPath"`
	ChildKeyPathValue any           `json:"childKeyPathValue" yaml:"childkeyPathValue"`
	ChildPermutations []Permutation `json:"childPermutations" yaml:"childPermutations"`
}

// Permutation structs are used to describe a single permutation
type Permutation struct {
	KeyPath                   []string       `json:"keyPath" yaml:"keyPath"`
	KeyPathValues             []any          `json:"keyPathValue" yaml:"keyPath"`
	KeyPathValueRelationships []Relationship `json:"keyPathValueRelationships" yaml:"KeyPathValueRelationships"`
}

// CreateRelationship is a constructor for the relationship struct
func CreateRelationship(parentValue any, childKeyPath []string, childKeyPathValue any, childPermutations []Permutation) Relationship {
	return Relationship{
		ParentValue:       parentValue,
		ChildKeyPath:      childKeyPath,
		ChildKeyPathValue: childKeyPathValue,
		ChildPermutations: childPermutations,
	}
}

// CreatePermutation is a constructor for the permutation struct
func CreatePermutation(keyPath []string, keyPathValues []any, keyPathValueRelationships []Relationship) Permutation {
	return Permutation{
		KeyPath:                   keyPath,
		KeyPathValues:             keyPathValues,
		KeyPathValueRelationships: keyPathValueRelationships,
	}
}

// Permute iterates over a list of permutation structs and permutes the base config with each of the permutations
func Permute(permutations []Permutation, baseConfig map[string]any) ([]map[string]any, error) {
	var configs []map[string]any
	var err error
	if len(permutations) == 0 {
		err = errors.New("no permutations provided")
		return configs, err
	}

	for _, keyPathValue := range permutations[0].KeyPathValues {
		permutedConfig, err := mapOperations.DeepCopyMap(baseConfig)
		if err != nil {
			return nil, err
		}

		permutedConfig, err = mapOperations.ReplaceValue(permutations[0].KeyPath, keyPathValue, permutedConfig)
		if err != nil {
			return nil, err
		}

		subPermutations := false
		for _, relationship := range permutations[0].KeyPathValueRelationships {
			if relationship.ParentValue == keyPathValue {
				if len(relationship.ChildKeyPath) > 1 && relationship.ChildKeyPathValue != nil {
					permutedConfig, err = mapOperations.ReplaceValue(relationship.ChildKeyPath, relationship.ChildKeyPathValue, permutedConfig)
					if err != nil {
						return nil, err
					}
				}

				var relationshipPermutedConfigs []map[string]any
				if len(relationship.ChildPermutations) > 0 {
					subPermutations = true
					relationshipPermutedConfigs, err = Permute(relationship.ChildPermutations, permutedConfig)
					if err != nil {
						return nil, err
					}
				}

				configs = append(configs, relationshipPermutedConfigs...)
			}
		}

		if !subPermutations {
			configs = append(configs, permutedConfig)
		}
	}

	var finalConfigs []map[string]any
	if len(permutations) == 1 {
		return configs, nil
	} else {
		for _, config := range configs {
			permutedConfigs, err := Permute(permutations[1:], config)
			if err != nil {
				return nil, err
			}

			finalConfigs = append(finalConfigs, permutedConfigs...)
		}
	}

	return finalConfigs, err
}
