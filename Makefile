PKG = github.com/airbnb/rudolph
VERSION := $(shell git describe --tags --always 2>/dev/null || echo dev)
TERRAFORM_DEPLOYMENTS_DIR = $(PWD)/deployments/environments
IMAGE ?= rudolph-api:local

.check-args:
ifndef ENV
	$(error ENV is not set. Use ENV=<YOUR_ENV> make plan, or export ENV.)
endif

TERRAFORM_DEPLOYMENTS_ENV_DIR = $(TERRAFORM_DEPLOYMENTS_DIR)/$(ENV)

deploy: .check-args test build tf-init
	terraform -chdir=$(TERRAFORM_DEPLOYMENTS_ENV_DIR) apply

plan: .check-args
	terraform -chdir=$(TERRAFORM_DEPLOYMENTS_ENV_DIR) plan

tf-init: .check-args
	terraform -chdir=$(TERRAFORM_DEPLOYMENTS_ENV_DIR) init

validate: .check-args
	terraform -chdir=$(TERRAFORM_DEPLOYMENTS_ENV_DIR) validate

destroy: .check-args tf-init
	terraform -chdir=$(TERRAFORM_DEPLOYMENTS_ENV_DIR) destroy

deps:
	go mod download
	go mod tidy

build: deps
	@"$(CURDIR)/scripts/build.sh"

image:
	docker build -t $(IMAGE) .

test:
	go test ./...

init-env: .check-args
	@"$(CURDIR)/scripts/new_env.sh" "$(ENV)"

.PHONY: .check-args deploy tf-init plan validate destroy deps build image test init-env
