#!/bin/bash

cd docker
docker-compose -f docker-compose.yml -f cypress-compose.yml  --env-file ".env.test" up
