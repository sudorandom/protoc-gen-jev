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
	@if [ ! -d ".venv" ]; then \
		echo "Creating Python virtualenv via uv..."; \
		uv venv .venv; \
	fi
	uv pip install -q -r examples/python/requirements.txt
	uv run python -c "import py_compile; py_compile.compile('examples/python/main.py', doraise=True)"
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
	uv run python examples/python/main.py
	@echo ""
	@echo "=== Running TypeScript Example ==="
	NODE_PATH="{{ justfile_directory() }}/examples/typescript/node_modules" npx --prefix examples/typescript tsx examples/typescript/index.ts
	@echo ""
	@echo "✔ All language examples completed successfully!"

# Setup local Laya environment in .venv-laya
setup-laya:
	@if [ ! -d ".venv-laya" ]; then \
		echo "Creating .venv-laya with Python 3.12..."; \
		uv venv .venv-laya --python 3.12; \
	fi
	@echo "Installing laya[serve]..."
	uv pip install --python .venv-laya/bin/python "laya[serve]"
	@echo "✔ Laya environment ready."

# Start local laya-serve in the foreground on port 8000
start-laya: setup-laya
	@echo "Starting Laya server on http://127.0.0.1:8000 ..."
	.venv-laya/bin/laya-serve --host 127.0.0.1 --port 8000

# Run multi-language examples targeting a running local Laya server
run-examples-laya: compile-examples
	@echo "Checking Laya server on http://127.0.0.1:8000 ..."
	@curl -sf http://127.0.0.1:8000/health >/dev/null || (echo "❌ Laya server is not running on :8000. Start it first with 'just start-laya'" && exit 1)
	@echo "✔ Laya server detected. Executing examples against Laya..."
	@echo ""
	@echo "=== Running Go Example against Laya ==="
	JEV_ENDPOINT="http://127.0.0.1:8000/v1/systemone" go run examples/go/main.go
	@echo ""
	@echo "=== Running Python Example against Laya ==="
	JEV_ENDPOINT="http://127.0.0.1:8000" uv run python examples/python/main.py
	@echo ""
	@echo "=== Running TypeScript Example against Laya ==="
	JEV_ENDPOINT="http://127.0.0.1:8000" NODE_PATH="{{ justfile_directory() }}/examples/typescript/node_modules" npx --prefix examples/typescript tsx examples/typescript/index.ts
	@echo ""
	@echo "✔ All language examples completed successfully against Laya!"


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
	golangci-lint run --fix ./...
	go fmt ./...


# ------------------------------------------------------------------------------
# Cleanup
# ------------------------------------------------------------------------------

# Clean build artifacts and generated files
clean:
	@echo "Cleaning artifacts..."
	rm -rf gen/ examples/gen/ protoc-gen-jev .venv/ .venv-laya/
	@echo "✔ Clean complete."
