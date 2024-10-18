package permutations

import (
	"fmt"

	mapOperations "github.com/rancher/shepherd/pkg/config/operations"
)

// ExamplePermute is an example of how to use the permutation objects and the permute function.
func ExamplePermute() {
	config := map[string]any{
		"foo1": map[string]any{
			"nested-foo1": []any{"bar1", "bar2"},
		},
		"foo2": []any{"bar3", "bar4"},
	}

	nestedFoo1Path := []string{"foo1", "nested-foo1"}
	nestedFoo1Value, _ := mapOperations.GetValue(nestedFoo1Path, config)
	foo1Permutation := CreatePermutation(nestedFoo1Path, nestedFoo1Value.([]any), []Relationship{})

	nestedFoo2Path := []string{"foo2"}
	nestedFoo2Value, _ := mapOperations.GetValue(nestedFoo2Path, config)
	foo2Permutation := CreatePermutation(nestedFoo2Path, nestedFoo2Value.([]any), []Relationship{})

	permutations := []Permutation{foo1Permutation, foo2Permutation}
	permutedConfigs, _ := Permute(permutations, config)

	for _, permutedConfig := range permutedConfigs {
		fmt.Println(permutedConfig)
		//Output:
		//map[foo1:map[nested-foo1:bar1] foo2:bar3]
		//map[foo1:map[nested-foo1:bar1] foo2:bar4]
		//map[foo1:map[nested-foo1:bar2] foo2:bar3]
		//map[foo1:map[nested-foo1:bar2] foo2:bar4]
	}

}
