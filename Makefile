# Build outputs go to ./bin. Go's build cache handles source-level
# staleness so these targets are phony — Make just orchestrates, the
# go toolchain decides what to rebuild.

BIN := bin

# All build targets. Add new commands here.
CMDS := slap demo

.PHONY: all $(CMDS) vet test clean

all: $(CMDS)

slap:
	go build -o $(BIN)/slap ./cmd/slap

demo:
	go build -o $(BIN)/demo ./cmd/demo

vet:
	go vet ./...

test:
	go test ./...

clean:
	rm -rf $(BIN)
	go clean -cache
