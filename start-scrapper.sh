#!/bin/bash

cleanup() {
  trap - SIGINT SIGTERM
  echo "stopping all services"
  echo "stopping dockers"
  docker compose -f devops/docker-compose.yml down
  kill 0
  exit 0
}
trap cleanup SIGINT SIGTERM

docker compose -f devops/docker-compose.yml up --wait
(cd backend && go run cmd/scrapper/main.go) &
wait
