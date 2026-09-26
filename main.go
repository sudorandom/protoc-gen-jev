// protoc-gen-jev is an experimental Protobuf compiler plugin for TypeSafe AI's Jev.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/sudorandom/protoc-gen-jev/internal/codegen"
	"github.com/sudorandom/protoc-gen-jev/internal/model"
	"github.com/sudorandom/protoc-gen-jev/internal/parser"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-version" || os.Args[1] == "-v") {
		fmt.Printf("protoc-gen-jev %s\n", version)
		return
	}

	var flags flag.FlagSet
	var targets []string
	addTarget := func(value string) error { targets = append(targets, value); return nil }
	flags.Func("target", "Target language (repeat for multiple targets)", addTarget)
	flags.Func("targets", "Alias for target", addTarget)

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		val := strings.Join(targets, ",")
		targetOptions, err := model.ParseTargets(val)
		if err != nil {
			return err
		}

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			if f.Desc.Package() == "jev.v1" {
				continue
			}

			var specs []model.MessageSpec
			for _, msg := range parser.AllMessages(f.Messages) {
				spec, err := parser.ProcessMessage(msg)
				if err != nil {
					return err
				}
				if len(spec.Questions) > 0 {
					specs = append(specs, spec)
				}
			}

			var serviceSpecs []model.ServiceSpec
			for _, svc := range f.Services {
				serviceSpec, err := parser.ProcessService(svc)
				if err != nil {
					return err
				}
				if len(serviceSpec.Methods) > 0 {
					serviceSpecs = append(serviceSpecs, serviceSpec)
				}
			}

			if len(specs) == 0 && len(serviceSpecs) == 0 {
				continue
			}

			// 1. Language-agnostic JSON spec
			if targetOptions.GenerateJSON {
				if err := codegen.GenerateJSON(gen, f, specs, serviceSpecs); err != nil {
					return err
				}
			}

			// 2. Native Go client
			if targetOptions.GenerateGo {
				codegen.GenerateGo(gen, f, specs, serviceSpecs)
			}

			// 3. TypeScript client
			if targetOptions.GenerateTypeScript {
				codegen.GenerateTypeScript(gen, f, specs, serviceSpecs)
			}

			// 4. Python client
			if targetOptions.GeneratePython {
				codegen.GeneratePython(gen, f, specs, serviceSpecs)
			}
		}

		return nil
	})
}
