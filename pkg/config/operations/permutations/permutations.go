package permutations

import (
	"errors"

	mapOperations "github.com/rancher/shepherd/pkg/config/operations"
)

// Relationship structs are used to create an association between a parent value and either a set of permutations or the value of a different key
type Relationship struct {
	ParentValue       any           `json:"parentValue" yaml:"parentValue"`
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

	//Apply the current permutation and all its relationships
	for _, keyPathValue := range permutations[0].KeyPathValues {
		permutedConfig, err := mapOperations.DeepCopyMap(baseConfig)
		if err != nil {
			return nil, err
		}

		permutedConfig, err = mapOperations.ReplaceValue(permutations[0].KeyPath, keyPathValue, permutedConfig)
		if err != nil {
			return nil, err
		}

		relationshipConfigs := []map[string]any{permutedConfig}
		if len(permutations[0].KeyPathValueRelationships) != 0 && permutations[0].KeyPathValueRelationships != nil {
			relationshipConfigs, err = applyRelationships(permutedConfig, permutations[0].KeyPathValueRelationships, keyPathValue)
			if err != nil {
				return nil, err
			}
		}

		configs = append(configs, relationshipConfigs...)
	}

	var finalConfigs []map[string]any
	if len(permutations) == 1 {
		return configs, nil
	} else {
		//Apply the rest of the permutations on top of the current permutation
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

// applyRelationships iterates over a list of relationships structs and applies them to a config
func applyRelationships(config map[string]any, relationships []Relationship, keyPathValue any) ([]map[string]any, error) {
	var err error
	var relationshipConfigs []map[string]any

	permutedConfigs := []map[string]any{config}
	if (len(relationships[0].ChildPermutations) > 0) && (relationships[0].ParentValue == keyPathValue) {
		//Apply the first relationship's permutations
		permutedConfigs, err = Permute(relationships[0].ChildPermutations, config)
		if err != nil {
			return nil, err
		}
	}

	relationshipConfigs = append(relationshipConfigs, permutedConfigs...)

	var finalConfigs []map[string]any
	if len(relationships) == 1 {
		return relationshipConfigs, nil
	} else {
		//Apply the rest of the relationships to all of the current configs
		for _, config := range relationshipConfigs {
			relationshipConfigs, err = applyRelationships(config, relationships[1:], keyPathValue)
			if err != nil {
				return nil, err
			}
			finalConfigs = append(finalConfigs, relationshipConfigs...)
		}
	}

	return finalConfigs, nil

}
