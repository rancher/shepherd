package defaults

import (
	"errors"
	"os"

	"sigs.k8s.io/yaml"
)

// LoadDefault is a helper to load objects that are stored in a yaml files into their go equivalent struct objects
func LoadDefault(defaultFile string, defaultName string, defaultObject interface{}) error {
	if defaultFile == "" {
		yaml.Unmarshal([]byte("{}"), defaultFile)
		err := errors.New("No default file found")
		return err
	}

	allString, err := os.ReadFile(defaultFile)
	if err != nil {
		panic(err)
	}

	var all map[string]map[string]interface{}
	err = yaml.Unmarshal(allString, &all)
	if err != nil {
		panic(err)
	}

	var keys []string
	for key := range all[defaultName] {
		keys = append(keys, key)
	}
	scoped := all[defaultName][keys[0]]
	scopedString, err := yaml.Marshal(scoped)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(scopedString, &defaultObject)
	if err != nil {
		panic(err)
	}

	return nil
}
