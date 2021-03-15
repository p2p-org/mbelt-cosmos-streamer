#!make
include .env
export $(shell sed 's/=.*//' .env)


reset:
	docker-compose stop && docker-compose rm -f && docker volume prune && ./start-docker.sh
init-ksql:
	docker-compose -f docker-compose.local.ksql.yml up -d
migration:
	docker run -v ${PWD}/migrations:/migrations --network host migrate/migrate -path=/migrations/ -database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/${POSTGRES_DB}\?sslmode=disable  up