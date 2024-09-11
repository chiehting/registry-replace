# 定義變量
GO_DEFINITION_DIR = ./definitions
GO_OUT_DIR = ./generated
APP_REGISTRY_REPLACE=registry-replace
K8S_YAML_FILE = k8s/registry-replace.yaml
DOCKER_COMPOSE_YAML = docker/docker-compose.yml

# tls 產生
tls: scripts/generate-tls.sh
	@bash scripts/generate-tls.sh cmd/tls

# generate 生成命令
generate: generate-api

.PHONY: generate-api
generate-api: definitions/$(APP_REGISTRY_REPLACE).api
	goctl api go --api=$(GO_DEFINITION_DIR)/$(APP_REGISTRY_REPLACE).api --dir="cmd" --style go_zero
# .PHONY: generate-rpc
# generate-rpc: definitions/$(APP_REGISTRY_REPLACE).proto
# 	goctl rpc protoc $(GO_DEFINITION_DIR)/$(APP_REGISTRY_REPLACE).proto --go_out=$(GO_OUT_DIR) --go-grpc_out=$(GO_OUT_DIR) --zrpc_out=rpc --style go_zero -m

.PHONY: buildImg
buildImg:
	docker-compose -f ${DOCKER_COMPOSE_YAML} build registry-replace-api
	printf "y\n" | docker system prune

.PHONY: apply
apply:
	kubectl apply -f ${K8S_YAML_FILE}

.PHONY: delete
delete:
	kubectl delete -f ${K8S_YAML_FILE}

.PHONY: run
run:
	cd cmd; go run registryreplace.go

