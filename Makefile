#
.PHONY: clean deploy healthz docker-clean

#
SHELL := /bin/bash

# 
APP := notification-service
PORT := 11000

ENVIRONMENT := $(shell echo ${bamboo_deploy_environment} | tr '[:upper:]' '[:lower:]')

format:
	go fmt ./...

clean: format
	go clean
	rm -f ./.deploy/bin/${APP}

build: clean
	@echo "===== Compile ====="
	GOOS=linux go build -o ./${APP} ./main.go

test-unit: build
	@echo "\t\t===== Unit Tests ====="
	go test -race -v `go list ./... | grep -v -e /vendor/ -e /mock/` | tee test-unit.txt
	# remove file
	@unlink test-unit.txt

test-coverage:
	go test -cover `go list ./... | grep -v -e /vendor/ -e /mock/`
	go test `go list ./... | grep -v -e /vendor/ -e /mock/` -coverprofile=cover.out
	go tool cover -func=cover.out
	# remove file
	@unlink cover.out

benchmark:
	go test -bench=. `go list ./... -race | grep -v -e /vendor/ -e /mock/`

healthz:
	smoke -f $(PWD)/.tests/smoke_test.yaml -u http://localhost:11000 -v

# DOCKER
docker-clean: 
	-docker rm -f ${APP}
	-docker network rm net-${APP}
	-docker rm -f zookeeper-${APP}
	-docker rm -f kafka-${APP}
	-docker rm -f prom-${APP}
	-docker rm -f graf-${APP}
	-docker rm -f httpbin-${APP}
	-docker network rm net-${APP}

docker-build:
	-docker rm -f ${APP}
	@if [ ! -z $${bamboo_SSH_FOLDER} ]; then \
		docker build -t mindgeek/${APP} --no-cache=true --build-arg private_key="$$(cat $${bamboo_SSH_FOLDER}/id_rsa)" .; \
	else \
		docker build -t mindgeek/${APP} --no-cache=true --network=host --build-arg ENV=local --build-arg private_key="$$(cat ${HOME}/.ssh/stash_id_rsa)" .; \
	fi

docker-network:
	-docker network create net-${APP}

docker-smoke-test:
	# remove previous
	-docker rm -f smoke-${APP}
	# run test
	docker run  --network net-${APP} \
	-v $$PWD/.tests/smoke_test.yaml:/etc/smoke/conf.d/smoke_test.yaml \
	--name smoke-${APP} bluehoodie/smoke \
	-f /etc/smoke/conf.d/smoke_test.yaml \
	-u http://${APP}:11000 -v

docker-kafka:
	# network
	-docker network create net-${APP}
	# zookeeper
	docker run -d --name zookeeper-${APP} \
	-e ZOOKEEPER_CLIENT_PORT=2181 \
    --network net-${APP} \
	-p 2181:2181 \
	confluentinc/cp-zookeeper
	# broker
	docker run -d --name kafka-${APP} \
	-e KAFKA_ZOOKEEPER_CONNECT=zookeeper-${APP}:2181 \
	-e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://kafka-${APP}:9092 \
	-e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
	--network net-${APP} \
	-p 9092:9092 \
	confluentinc/cp-kafka
	# httpbin
	docker run -d -p 80:80 --name httpbin-${APP} kennethreitz/httpbin
	#
	sleep 20

docker-launch:
	docker run --name ${APP} -d -p 11000:11000 -v $$(pwd)/.deploy:/data --network net-${APP} \
	-v $$PWD/.tests/conf.d/notification-service.toml:/etc/notification-service/conf.d/notification-service.toml \
	mindgeek/${APP}

docker-monitor:
	# prometheus
	docker run --name prom-${APP} -d -p 9090:9090 -v ${PWD}/.monitoring/prometheus/:/etc/prometheus/ \
	--network net-${APP} \
	prom/prometheus
	# grafana
	docker run --name graf-${APP} -d -p 3000:3000 \
	-v  ${PWD}/.monitoring/grafana/provisioning/:/etc/grafana/provisioning/ \
	--network net-${APP} \
	-e GF_SECURITY_ADMIN_PASSWORD=foobar \
	-e GF_USERS_ALLOW_SIGN_UP=false \
	grafana/grafana

docker-structure-test:
	docker run -i --rm -v ${PWD}/:/src/:ro -v /var/run/docker.sock:/var/run/docker.sock  gcr.io/gcp-runtimes/container-structure-test test --image mindgeek/${APP} --config /src/container-structure-test.yaml

docker-copy-artifact:
	docker exec ${APP} sh -c "mkdir /data/bin/"
	docker exec ${APP} sh -c "cp /bin/${APP} /data/bin/${APP}"
	docker exec ${APP} sh -c "chmod 777 -R /data/bin/${APP}"


artifact: docker-build docker-network docker-kafka docker-launch docker-smoke-test docker-copy-artifact docker-clean
	@echo "Artifact build completed";

pre-commit:  build docker-build docker-network docker-kafka docker-launch docker-smoke-test docker-clean
	@echo "Pre-commit completed";

docker-prometheus:


deploy: # https://stash.mgcorp.co/projects/LT/repos/bamboo.specs/browse/bamboo-specs/src/main/java/com/mindgeek/ads/TJDSP.java
	echo ${ENVIRONMENT}
	echo ${DEPLOY_ENVIRONMENT}
	echo ${DESTINATION_SERVERS}
ifneq (,$(findstring uat,${ENVIRONMENT}))
	$(MAKE) deploy-uat	
else
	$(MAKE) deploy-production	
endif

deploy-uat:
	@for S in $(DESTINATION_SERVERS); do \
		ssh -i ${bamboo_SSH_FOLDER}/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o GlobalKnownHostsFile=/dev/null adbroker@$$S "supervisorctl stop ${APP}_${ENVIRONMENT}:"; \
		rsync --stats --progress \
			-clrv -e 'ssh -i ${bamboo_SSH_FOLDER}/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o GlobalKnownHostsFile=/dev/null ' \
			--verbose --delete-after --ignore-errors --force \
			--exclude=".ssh" --safe-links --checksum  \
			-avz ./.deploy/bin/* \
			adbroker@$$S:/home/adbroker/deep-dsp/${ENVIRONMENT}/apps/${APP}/; \
		ssh -i ${bamboo_SSH_FOLDER}/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o GlobalKnownHostsFile=/dev/null adbroker@$$S "supervisorctl start ${APP}_${ENVIRONMENT}:"; \
	done;

deploy-production:
	@echo production
# deploy primary
	@for S in $(DESTINATION_SERVERS); do \
		ssh -i ${bamboo_SSH_FOLDER}/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o GlobalKnownHostsFile=/dev/null adbroker@$$S "mkdir -p /home/adbroker/deploy/${APP}/"; \
		rsync --stats --progress \
			-clrv -e 'ssh -i ${bamboo_SSH_FOLDER}/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o GlobalKnownHostsFile=/dev/null ' \
			--verbose --delete-after --ignore-errors --force \
			--exclude=".ssh" --safe-links --checksum  \
			-avz ./.deploy/bin/* \
			adbroker@$$S:/home/adbroker/deploy/${APP}/; \
#		ssh -i ${bamboo_SSH_FOLDER}/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o GlobalKnownHostsFile=/dev/null adbroker@$$S "supervisorctl stop ${APP}: && rm /home/adbroker/apps/${APP}/${APP} && cp /home/adbroker/deploy/${APP}/${APP} /home/adbroker/apps/${APP}/${APP}"; \
#		ssh -i ${bamboo_SSH_FOLDER}/id_rsa -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o GlobalKnownHostsFile=/dev/null adbroker@$$S "supervisorctl start ${APP}:"; \
	done;