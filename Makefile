.PHONY: *

test:
	docker compose -f ./deploy/local/integration_test/docker-compose.yml down --remove-orphans
	docker compose -f ./deploy/local/integration_test/docker-compose.yml up --build --exit-code-from=integration-test

run:
	-docker compose -f ./deploy/local/run/docker-compose-local.yml down --remove-orphans
	docker compose -f ./deploy/local/run/docker-compose-local.yml up --build --attach=server --attach=wa-worker --attach=vacancy-worker

# command for internal
run-internal:
	-docker compose -f ./deploy/local/run/docker-compose.yml down --remove-orphans
	docker compose -f ./deploy/local/run/docker-compose.yml up --build --attach=server --attach=wa-worker --attach=vacancy-worker

## command for Batha server which run old AMI Linux version. In there there is no "docker compose".
deploy-ec2:
	-make stop-ec2
	docker-compose -f ./deploy/aws/ec2/docker-compose.yml up --build -d

logs-ec2:
	docker-compose -f ./deploy/aws/ec2/docker-compose.yml logs -f

stop-ec2:
	docker-compose -f ./deploy/aws/ec2/docker-compose.yml down --remove-orphans	
