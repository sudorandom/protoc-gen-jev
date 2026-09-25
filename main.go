// protoc-gen-jev is an experimental Protobuf compiler plugin for TypeSafe AI's Jev.
package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"protoc-gen-jev/internal/codegen"
	"protoc-gen-jev/internal/model"
	"protoc-gen-jev/internal/parser"
)

func main() {
	var flags flag.FlagSet
	targetOpt := flags.String("target", "all", "Target languages to generate: go, ts, python, json, all")
	targetsOpt := flags.String("targets", "", "Alias for target")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		val := *targetOpt
		if *targetsOpt != "" {
			val = *targetsOpt
		}
		targetOptions := model.ParseTargets(val)

		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			var specs []model.MessageSpec
			for _, msg := range f.Messages {
				spec := parser.ProcessMessage(msg)
				if len(spec.Questions) > 0 {
					specs = append(specs, spec)
				}
			}

			if len(specs) == 0 {
				continue
			}

			// 1. Language-agnostic JSON spec
			if targetOptions.GenerateJSON {
				if err := codegen.GenerateJSON(gen, f, specs); err != nil {
					return err
				}
			}

			// 2. Native Go client
			if targetOptions.GenerateGo {
				codegen.GenerateGo(gen, f, specs)
			}

			// 3. TypeScript client
			if targetOptions.GenerateTypeScript {
				codegen.GenerateTypeScript(gen, f, specs)
			}

			// 4. Python client
			if targetOptions.GeneratePython {
				codegen.GeneratePython(gen, f, specs)
			}
		}

		return nil
	})
}
