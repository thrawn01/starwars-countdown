#!/bin/sh

DOCKER_REPO=thrawn01
docker build -t ${DOCKER_REPO}/starwars-countdown:latest .
docker push ${DOCKER_REPO}/starwars-countdown:latest
