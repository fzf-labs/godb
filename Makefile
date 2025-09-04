SHELL = /bin/bash
GitBranch = $(shell git rev-parse --abbrev-ref HEAD)

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

.PHONY: lint
# make lint  golang使用最多的第三方静态程序分析工具
lint:
	@golangci-lint run ./... -v

.PHONY: git-clean
# make git-clean  git clean
git-clean:
	#清除开始
	@git checkout --orphan latest_branch
	@git add -A
	@git commit -am "clean"
	@git branch -D ${GitBranch}
	@git branch -m ${GitBranch}
	@git push -f origin ${GitBranch}
	#清除结束

# 创建新的 tag
git-new-tag:
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

