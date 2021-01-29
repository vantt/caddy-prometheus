#!/bin/bash
docker-compose -f docker-compose.yml --env-file ".test.env" up
