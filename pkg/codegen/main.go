package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"

	fleet "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/norman/types"
	catalogv1 "github.com/rancher/rancher/pkg/apis/catalog.cattle.io/v1"
	provisioningv1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	rkev1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	"github.com/rancher/shepherd/pkg/codegen/generator"
	"github.com/rancher/shepherd/pkg/schemas/factory"
	managementSchema "github.com/rancher/shepherd/pkg/schemas/management.cattle.io/v3"
	planv1 "github.com/rancher/system-upgrade-controller/pkg/apis/upgrade.cattle.io/v1"
	controllergen "github.com/rancher/wrangler/v3/pkg/controller-gen"
	"github.com/rancher/wrangler/v3/pkg/controller-gen/args"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

// main initializes the code generation for controllers and clients.
// It generates clients for various API groups using controllergen.Run.
// It also generates clients for specific schemas using generator.GenerateClient.
// Then, it calls replaceClientBasePackages to replace imports in the generated clients.
// Finally, it replaces imports and adds controller test session for generated
func main() {
	err := os.Unsetenv("GOPATH")
	if err != nil {
		return
	}

	generatedControllerPaths := map[string]string{
		"AppsControllerPath":               "./pkg/generated/controllers/apps",
		"CoreControllerPath":               "./pkg/generated/controllers/core",
		"RBACControllerPath":               "./pkg/generated/controllers/rbac",
		"BatchControllerPath":              "./pkg/generated/controllers/batch",
		"ManagementControllerPath":         "./pkg/generated/controllers/management.cattle.io",
		"ClusterCattleControllerPath":      "./pkg/generated/controllers/cluster.cattle.io",
		"CatalogCattleControllerPath":      "./pkg/generated/controllers/catalog.cattle.io",
		"UpgradeCattleControllerPath":      "./pkg/generated/controllers/upgrade.cattle.io",
		"ProvisioningCattleControllerPath": "./pkg/generated/controllers/provisioning.cattle.io",
		"FleetCattleControllerPath":        "./pkg/generated/controllers/fleet.cattle.io",
		"RKECattleControllerPath":          "./pkg/generated/controllers/rke.cattle.io",
		"ClusterXK8sControllerPath":        "./pkg/generated/controllers/cluster.x-k8s.io",
		"ExtCattleControllerPath":          "./pkg/generated/controllers/ext.cattle.io",
	}

	controllergen.Run(args.Options{
		OutputPackage: "github.com/rancher/shepherd/pkg/generated",
		Boilerplate:   "pkg/codegen/boilerplate.go.txt",
		Groups: map[string]args.Group{
			appsv1.GroupName: {
				Types: []interface{}{
					appsv1.ControllerRevision{},
					appsv1.Deployment{},
					appsv1.DaemonSet{},
					appsv1.ReplicaSet{},
					appsv1.StatefulSet{},
				},
			},
			corev1.GroupName: {
				Types: []interface{}{
					corev1.Event{},
					corev1.Node{},
					corev1.Namespace{},
					corev1.LimitRange{},
					corev1.ResourceQuota{},
					corev1.Secret{},
					corev1.Service{},
					corev1.ServiceAccount{},
					corev1.Endpoints{},
					corev1.ConfigMap{},
					corev1.PersistentVolume{},
					corev1.PersistentVolumeClaim{},
					corev1.Pod{},
				},
			},
			rbacv1.GroupName: {
				Types: []interface{}{
					rbacv1.Role{},
					rbacv1.RoleBinding{},
					rbacv1.ClusterRole{},
					rbacv1.ClusterRoleBinding{},
				},
				OutputControllerPackageName: "rbac",
			},
			batchv1.GroupName: {
				Types: []interface{}{
					batchv1.Job{},
					batchv1.CronJob{},
				},
			},
			"management.cattle.io": {
				PackageName: "management.cattle.io",
				Types: []interface{}{
					// All structs with an embedded ObjectMeta field will be picked up
					"./vendor/github.com/rancher/rancher/pkg/apis/management.cattle.io/v3",
				},
			},
			"cluster.cattle.io": {
				PackageName: "cluster.cattle.io",
				Types: []interface{}{
					// All structs with an embedded ObjectMeta field will be picked up
					"./vendor/github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3",
				},
			},
			"ext.cattle.io": {
				PackageName: "ext.cattle.io",
				Types: []interface{}{
					// All structs with an embedded ObjectMeta field will be picked up
					"./vendor/github.com/rancher/rancher/pkg/apis/ext.cattle.io/v1",
				},
			},
			"catalog.cattle.io": {
				PackageName: "catalog.cattle.io",
				Types: []interface{}{
					catalogv1.App{},
					catalogv1.ClusterRepo{},
					catalogv1.Operation{},
				},
				GenerateClients: true,
			},
			"upgrade.cattle.io": {
				PackageName: "upgrade.cattle.io",
				Types: []interface{}{
					planv1.Plan{},
				},
				GenerateClients: true,
			},
			"provisioning.cattle.io": {
				Types: []interface{}{
					provisioningv1.Cluster{},
				},
				GenerateClients: true,
			},
			"fleet.cattle.io": {
				Types: []interface{}{
					fleet.Bundle{},
					fleet.Cluster{},
					fleet.ClusterGroup{},
				},
			},
			"rke.cattle.io": {
				Types: []interface{}{
					rkev1.RKEBootstrap{},
					rkev1.RKEBootstrapTemplate{},
					rkev1.RKECluster{},
					rkev1.RKEControlPlane{},
					rkev1.ETCDSnapshot{},
					rkev1.CustomMachine{},
				},
				GenerateClients: true,
			},
			"cluster.x-k8s.io": {
				Types: []interface{}{
					capi.Machine{},
					capi.MachineSet{},
					capi.MachineDeployment{},
					capi.Cluster{},
				},
			},
		},
	})

	clusterAPIVersion := &types.APIVersion{Group: capi.GroupVersion.Group, Version: capi.GroupVersion.Version, Path: "/v1"}
	generator.GenerateClient(factory.Schemas(clusterAPIVersion).Init(func(schemas *types.Schemas) *types.Schemas {
		return schemas.MustImportAndCustomize(clusterAPIVersion, capi.Machine{}, func(schema *types.Schema) {
			schema.ID = "cluster.x-k8s.io.machine"
		})
	}), nil)

	generator.GenerateClient(managementSchema.Schemas, map[string]bool{
		"userAttribute": true,
	})

	if err := replaceClientBasePackages(); err != nil {
		panic(err)
	}

	// Loop through all generated controller paths and replace imports and
	// and add test session
	for _, path := range generatedControllerPaths {
		if err := replaceImports(path); err != nil {
			panic(err)
		}

		if err := addControllerTestSession(path); err != nil {
			panic(err)
		}
	}
}

// replaceClientBasePackages walks through the zz_generated_client generated by generator.GenerateClient to replace imports from
// "github.com/rancher/norman/clientbase" to "github.com/rancher/shepherd/pkg/clientbase" to use our modified code of the
// session.Session tracking the resources created by the Management Client.
func replaceClientBasePackages() error {
	return filepath.Walk("./clients/rancher/generated", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasPrefix(info.Name(), "zz_generated_client") {
			input, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			replacement := bytes.Replace(input, []byte("github.com/rancher/norman/clientbase"), []byte("github.com/rancher/shepherd/pkg/clientbase"), -1)

			if err = os.WriteFile(path, replacement, 0666); err != nil {
				return err
			}
		}

		return nil
	})
}

// Walk through the generated controllers and add test session
// to necessary functions and structs
func addControllerTestSession(root string) error {
	err := filepath.Walk(root, processInterfaceFile)
	if err != nil {
		return err
	}
	return nil
}

// replaceImports walks through the specified directory and replaces certain imports in Go files.
// It replaces the import "github.com/rancher/wrangler/v3/pkg/generic" with "github.com/rancher/shepherd/pkg/wrangler/pkg/generic"
// in all files ending with ".go".
// It also replaces specific function calls in files starting with "factory", and "interface".
// The replaced function calls are:
// - "New(c.ControllerFactory())" with "New(c.ControllerFactory(), c.Opts.TS)"
// - "controller.NewSharedControllerFactoryWithAgent(userAgent, c.ControllerFactory())" with "controller.NewSharedControllerFactoryWithAgent(userAgent, c.ControllerFactory()), c.Opts.TS"
// - "controller.SharedControllerFactory)" with "controller.SharedControllerFactory, ts *session.Session)"
// - "g.controllerFactory)" with "g.controllerFactory, g.ts)"
// - "v.controllerFactory)" with "v.controllerFactory, v.ts)"
// The function returns an error if there was a problem reading or writing files.
func replaceImports(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var replacement []byte

		if strings.HasSuffix(info.Name(), "go") {
			input, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			replacement = bytes.Replace(input, []byte("github.com/rancher/wrangler/v3/pkg/generic"), []byte("github.com/rancher/shepherd/pkg/wrangler/pkg/generic"), -1)
			if err = os.WriteFile(path, replacement, 0666); err != nil {
				return err
			}
		}

		if strings.HasPrefix(info.Name(), "factory") {
			input, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			replacement = bytes.Replace(input, []byte("New(c.ControllerFactory())"), []byte("New(c.ControllerFactory(), c.Opts.TS)"), -1)
			if err = os.WriteFile(path, replacement, 0666); err != nil {
				return err
			}
		}

		if strings.HasPrefix(info.Name(), "factory") {
			input, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			replacement = bytes.Replace(input, []byte("controller.NewSharedControllerFactoryWithAgent(userAgent, c.ControllerFactory())"), []byte("controller.NewSharedControllerFactoryWithAgent(userAgent, c.ControllerFactory()), c.Opts.TS"), -1)
			if err = os.WriteFile(path, replacement, 0666); err != nil {
				return err
			}
		}

		if strings.HasPrefix(info.Name(), "interface") {
			input, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			replacement = bytes.Replace(input, []byte("controller.SharedControllerFactory)"), []byte("controller.SharedControllerFactory, ts *session.Session)"), -1)
			if err = os.WriteFile(path, replacement, 0666); err != nil {
				return err
			}
			input, err = os.ReadFile(path)
			if err != nil {
				return err
			}
			replacement = bytes.Replace(input, []byte("g.controllerFactory)"), []byte("g.controllerFactory, g.ts)"), -1)
			if err = os.WriteFile(path, replacement, 0666); err != nil {
				return err
			}
			input, err = os.ReadFile(path)
			if err != nil {
				return err
			}
			replacement = bytes.Replace(input, []byte("v.controllerFactory)"), []byte("v.controllerFactory, v.ts)"), -1)
			if err = os.WriteFile(path, replacement, 0666); err != nil {
				return err
			}
		}
		return nil
	})
}

// Check if import already exists
func addImport(fset *token.FileSet, filename string, importPath string) error {
	if importPath == "" {
		return errors.New("empty import path")
	}

	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Check if session import is there and do nothing
	for _, i := range node.Imports {
		if i.Path.Value == importPath {
			println("Import already included in file:", filename)
			return nil
		}
	}

	// Create a new import spec
	newImport := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: importPath,
		},
	}

	// Insert the new import spec in the right place
	found := false
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			genDecl.Specs = append(genDecl.Specs, newImport)
			found = true
			break
		}
	}

	// If no import declaration was found, create a new one
	if !found {
		node.Decls = append([]ast.Decl{
			&ast.GenDecl{
				Tok: token.IMPORT,
				Specs: []ast.Spec{
					newImport,
				},
			},
		}, node.Decls...)
	}

	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, node)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, buf.Bytes(), 0666)
	if err != nil {
		return err
	}
	return nil
}

// processInterfaceFile processes the specified file and adds import, new struct line,
// and new function line to the specified blocks within the file.
func processInterfaceFile(path string, info os.FileInfo, err error) error {
	const importPath = `"github.com/rancher/shepherd/pkg/session"`
	const newStructLine = "\tts                *session.Session"
	const newFuncLine = "\t\tts:                ts,"

	if !info.IsDir() && strings.HasSuffix(info.Name(), "interface.go") {
		fset := token.NewFileSet()
		err := addImport(fset, path, importPath)
		if err != nil {
			return err
		}

		err = appendNewlineToBlockInFile(path, "group", newStructLine)
		if err != nil {
			return err
		}

		err = appendNewlineToBlockInFile(path, "&group", newFuncLine)
		if err != nil {
			return err
		}

		err = appendNewlineToBlockInFile(path, "version", newStructLine)
		if err != nil {
			return err
		}

		err = appendNewlineToBlockInFile(path, "&version", newFuncLine)
		if err != nil {
			return err
		}
	}
	return nil
}

// appendNewlineToBlockInFile takes a path to a file, a code block(struct,return) in a file to update
// and the string to insert in a new line within the block
func appendNewlineToBlockInFile(filePath, blockName, newLine string) error {
	// Read the file contents
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Convert content to string and split into lines
	input := string(content)
	lines := strings.Split(input, "\n")
	blockStart := -1
	blockEnd := -1
	braceCount := 0

	// Create a regex for matching function declaration
	var re *regexp.Regexp
	re = regexp.MustCompile(fmt.Sprintf(`(?m)(type\s+%s\s+struct|\s*return\s*%s)\s*{\s*`, blockName, blockName))

	// Find the start and end of the specified function
	for i, line := range lines {
		if blockStart == -1 {
			if re.MatchString(line) {
				blockStart = i
				braceCount = 1
			}
		} else {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			if braceCount == 0 {
				blockEnd = i
				break
			}
		}
	}

	// If the target is found, insert the new line before the closing brace
	if blockStart != -1 && blockEnd != -1 {
		// Find the last non-empty line before the closing brace
		insertPos := blockEnd
		for i := blockEnd - 1; i > blockStart; i-- {
			if strings.TrimSpace(lines[i]) != "" {
				insertPos = i + 1
				break
			}
		}

		// Insert the new line
		lines = append(lines[:insertPos], append([]string{newLine}, lines[insertPos:]...)...)

		// Join the lines back together
		modifiedContent := strings.Join(lines, "\n")

		// Write the modified content back to the file
		err = os.WriteFile(filePath, []byte(modifiedContent), 0644)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}

	return nil
}
