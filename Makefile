reset:
	docker-compose stop && docker-compose rm -f && docker volume prune && ./start-docker.sh