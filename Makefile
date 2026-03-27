APP      := demo
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
GOPATH   := $(shell go env GOPATH)
GOBIN    := $(GOPATH)/bin
GOOS     := $(shell go env GOOS)
GOARCH   := $(shell go env GOARCH)

API_PROTO_FILES := $(shell find api -name "*.proto")

.PHONY: all
all: wire build

# ==============================
#  依赖工具安装
# ==============================

.PHONY: install-tools
install-tools:
	@echo ">> installing protoc plugins"
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# ==============================
#  代码生成
# ==============================

.PHONY: proto
proto:
	@echo ">> generating proto files"
	@for f in $(API_PROTO_FILES); do \
		protoc \
			--proto_path=. \
			--proto_path=./third_party \
			--go_out=paths=source_relative:. \
			--go-http_out=paths=source_relative:. \
			--go-grpc_out=paths=source_relative:. \
			--go-errors_out=paths=source_relative:. \
			$$f; \
		echo "  generated $$f"; \
	done

.PHONY: wire
wire:
	@echo ">> running wire"
	cd cmd/server && $(GOBIN)/wire || go run github.com/google/wire/cmd/wire@latest

# ==============================
#  构建
# ==============================

.PHONY: build
build:
	@echo ">> building $(APP) ($(GOOS)/$(GOARCH))"
	mkdir -p bin
	go build -buildvcs=false -ldflags "\
		-X main.Name=$(APP) \
		-X main.Version=$(VERSION)" \
		-o bin/$(APP) ./cmd/server

.PHONY: build-linux
build-linux:
	@echo ">> cross-compiling for linux/amd64"
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false \
		-ldflags "-X main.Name=$(APP) -X main.Version=$(VERSION)" \
		-o bin/$(APP)-linux-amd64 ./cmd/server

# ==============================
#  运行
# ==============================

.PHONY: run
run:
	@echo ">> running $(APP)"
	go run -buildvcs=false ./cmd/server -conf ./configs

# ==============================
#  测试
# ==============================

.PHONY: test
test:
	@echo ">> running tests"
	go test -buildvcs=false -v -race -cover ./...

.PHONY: test-coverage
test-coverage:
	@echo ">> running tests with coverage report"
	go test -buildvcs=false -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "  coverage report: coverage.html"

# ==============================
#  代码质量
# ==============================

.PHONY: lint
lint:
	@echo ">> running golangci-lint"
	$(GOBIN)/golangci-lint run ./...

.PHONY: vet
vet:
	@echo ">> running go vet"
	go vet ./...

.PHONY: fmt
fmt:
	@echo ">> formatting code"
	gofmt -s -w .
	goimports -w . 2>/dev/null || true

# ==============================
#  依赖管理
# ==============================

.PHONY: tidy
tidy:
	@echo ">> tidying modules"
	go mod tidy

.PHONY: vendor
vendor:
	@echo ">> vendoring modules"
	go mod vendor

# ==============================
#  Docker
# ==============================

.PHONY: docker-build
docker-build:
	@echo ">> building docker image $(APP):$(VERSION)"
	docker build \
		--build-arg APP=$(APP) \
		--build-arg VERSION=$(VERSION) \
		-t $(APP):$(VERSION) \
		-t $(APP):latest \
		-f Dockerfile .

.PHONY: docker-run
docker-run:
	@echo ">> running docker container"
	docker run --rm -p 8000:8000 -p 9000:9000 $(APP):latest

# ==============================
#  清理
# ==============================

.PHONY: clean
clean:
	@echo ">> cleaning"
	rm -rf bin/ coverage.out coverage.html

.PHONY: help
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "代码生成:"
	@echo "  proto          根据 .proto 文件生成 Go 代码"
	@echo "  wire           生成 Wire 依赖注入代码"
	@echo ""
	@echo "构建与运行:"
	@echo "  build          编译为本地二进制 (bin/$(APP))"
	@echo "  build-linux    交叉编译 linux/amd64"
	@echo "  run            go run 直接运行"
	@echo ""
	@echo "测试与质量:"
	@echo "  test           运行所有测试"
	@echo "  test-coverage  生成覆盖率报告 (coverage.html)"
	@echo "  lint           golangci-lint 静态检查"
	@echo "  vet            go vet 检查"
	@echo "  fmt            格式化代码"
	@echo ""
	@echo "依赖管理:"
	@echo "  tidy           go mod tidy"
	@echo "  vendor         go mod vendor"
	@echo "  install-tools  安装所有开发工具"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build   构建镜像 $(APP):$(VERSION)"
	@echo "  docker-run     运行容器"
	@echo ""
	@echo "其他:"
	@echo "  clean          删除构建产物"
	@echo "  all            wire + build (默认)"
	@echo ""
