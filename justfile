# Justfile for protoc-gen-jev
set shell := ["bash", "-c"]
export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE := env_var_or_default("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")

# Default recipe: lint and run all tests
default: lint test

# ------------------------------------------------------------------------------
# Build & Generation
# ------------------------------------------------------------------------------

# Compile the protoc-gen-jev binary
build:
	@echo "Building protoc-gen-jev..."
	go build -o protoc-gen-jev .
	@echo "✔ Binary built: ./protoc-gen-jev"

# Generate code using buf
generate: build
	@echo "Running buf generate for testdata..."
	buf generate
	@echo "Running buf generate for examples..."
	buf generate --template examples/buf.gen.yaml examples/proto
	@echo "✔ Code generation complete."

# ------------------------------------------------------------------------------
# Testing & Golden Files
# ------------------------------------------------------------------------------

# Run syntax and type checks across all generated languages
test-syntax: generate
	@echo "Checking generated Go code..."
	go vet ./gen/jev/...
	@echo "Checking generated Python syntax..."
	python -c "import py_compile, glob; [py_compile.compile(f, doraise=True) for f in glob.glob('gen/jev/**/*.py', recursive=True)]"
	@echo "Checking generated TypeScript type signatures..."
	npx --package typescript tsc --noEmit --skipLibCheck testdata/types/@typesafe-ai/sdk/index.d.ts `find gen/jev -name "*.ts"`
	@echo "✔ Generated Go, Python, and TypeScript code verified."

# Compile and verify all examples without executing them
compile-examples: generate
	@echo "Compiling Go example..."
	go build -o /dev/null ./examples/go
	@echo "Compiling Python example..."
	python -c "import py_compile; py_compile.compile('examples/python/main.py', doraise=True)"
	@echo "Compiling TypeScript example..."
	@if [ ! -d "examples/typescript/node_modules" ]; then \
		echo "Installing TypeScript example dependencies..."; \
		npm --prefix examples/typescript install --no-audit; \
	fi
	npx --prefix examples/typescript tsc --noEmit -p examples/typescript/tsconfig.json
	@echo "✔ All examples compiled successfully."

# Run all unit tests, syntax checks, golden file verification, and compile examples
test: generate test-syntax compile-examples
	@echo "Running Go tests..."
	go test -v ./internal/... .

# Run real end-to-end examples across all languages (Go, Python, TypeScript)
run-examples: compile-examples
	@echo "=== Running Go Example ==="
	go run examples/go/main.go
	@echo ""
	@echo "=== Running Python Example ==="
	python -c "import typesafe_sdk" 2>/dev/null || pip install -q typesafe-sdk
	python examples/python/main.py
	@echo ""
	@echo "=== Running TypeScript Example ==="
	NODE_PATH="{{ justfile_directory() }}/examples/typescript/node_modules" npx --prefix examples/typescript tsx examples/typescript/index.ts
	@echo ""
	@echo "✔ All language examples completed successfully!"

# Update golden files with newly generated files
update-golden: generate
	@echo "Updating golden files..."
	mkdir -p testdata/golden
	rm -rf testdata/golden/*
	cp -r gen/jev/* testdata/golden/
	@echo "✔ Golden files updated in testdata/golden"

# Delete existing golden files
clean-golden:
	@echo "Deleting golden files..."
	rm -rf testdata/golden
	@echo "✔ Golden files removed."

# ------------------------------------------------------------------------------
# Linting & Formatting
# ------------------------------------------------------------------------------

# Run all linters
lint: generate
	@echo "Linting protobuf files..."
	buf lint
	@echo "Running go vet..."
	go vet ./internal/... ./pkg/... ./examples/... .
	@echo "Running golangci-lint..."
	golangci-lint run ./...

# Run goreleaser to create a release or check configuration
release *args="release --clean":
	goreleaser {{ args }}

# Format all Go code
format:
	@echo "Formatting Go code..."
	go fmt ./...

# ------------------------------------------------------------------------------
# Cleanup
# ------------------------------------------------------------------------------

# Clean build artifacts and generated files
clean:
	@echo "Cleaning artifacts..."
	rm -rf gen/ examples/gen/ protoc-gen-jev
	@echo "✔ Clean complete."
