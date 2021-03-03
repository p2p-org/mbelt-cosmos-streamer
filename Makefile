reset:
	docker-compose stop && docker-compose rm -f && docker volume prune && ./start-docker.sh
init-ksql:
	docker-compose -f docker-compose.local.ksql.yml up -d