SHELL = /bin/bash
TEST_PKGS := ./...
COVER_PKGS := $(shell go list ./... | grep -v '/orm/example/gorm/postgres/gorm_gen_')

.PHONY: mod
# make mod  golang库更新
mod:
	#环境更新-开始
	@go mod download
	@go mod tidy
	#环境更新-结束

.PHONY: fmt
# make fmt  格式化代码
fmt:
	@gofmt -s -w .

.PHONY: gci
# make gci 格式化代码引入
gci:
	@gci write ./...

.PHONY: vet
# make vet golang官方命令,用于检查代码中的问题.
vet:
	@go vet ./...

.PHONY: test
# make test 运行全量测试
test:
	@go test $(TEST_PKGS)

.PHONY: cover
# make cover 运行覆盖率并排除生成示例包
cover:
	@go test $(COVER_PKGS) -coverprofile=/tmp/godb.cover
	@go tool cover -func=/tmp/godb.cover | tail -n 1

.PHONY: bootstrap-postgres
# make bootstrap-postgres 准备本地 PostgreSQL 测试数据库
bootstrap-postgres:
	@scripts/ci/bootstrap-postgres.sh

.PHONY: comments
# make comments 检查导出 API 注释
comments:
	@go run ./scripts/check_exported_comments.go

.PHONY: ci
# make ci 运行格式检查、注释检查、vet、测试和覆盖率
ci:
	@test -z "$$(gofmt -l .)"
	@$(MAKE) lint
	@$(MAKE) comments
	@$(MAKE) vet
	@$(MAKE) test
	@$(MAKE) cover

.PHONY: release-snapshot
# make release-snapshot 预览发布产物
release-snapshot:
	@go run github.com/goreleaser/goreleaser/v2@v2.16.0 release --snapshot --clean --skip=publish

.PHONY: release-tag
# make release-tag 创建并推送下一个 patch tag
release-tag:
	# 获取当前最新的 tag
	$(eval LATEST_TAG := $(shell git tag -l | sort -V | tail -n 1))
	@echo "当前最新的 tag: ${LATEST_TAG}"
	# 检查当前代码是否比最新 tag 更新
	@if [ "$$(git rev-list ${LATEST_TAG}..HEAD --count)" -eq "0" ]; then \
		echo "❌ 当前代码与最新 tag ${LATEST_TAG} 相同，无需创建新版本"; \
		exit 1; \
	fi
	@echo "✅ 检测到 $$(git rev-list ${LATEST_TAG}..HEAD --count) 个新提交，可以创建新版本"
	# 生成新的版本号
	$(eval NEW_VERSION := $(shell echo ${LATEST_TAG} | awk -F. '{$$NF=$$NF+1; print $$1"."$$2"."$$NF}'))
	@echo "新的版本号: ${NEW_VERSION}"
	# 创建新的 tag
	@git tag -a ${NEW_VERSION} -m "release ${NEW_VERSION}"
	# 推送新的 tag
	@git push origin ${NEW_VERSION}

.PHONY: lint
# make lint  golang使用最多的第三方静态程序分析工具
lint:
	@golangci-lint run ./... -v

# 创建新的 tag
git-new-tag: release-tag

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
