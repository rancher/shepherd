//go:generate go mod vendor
//go:generate go run pkg/codegen/generator/cleanup/main.go -mod vendor
//go:generate go run pkg/codegen/main.go -mod vendor
//go:generate rm -rf vendor
//go:generate go fmt ./...

package main
