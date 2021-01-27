# Using Cypress
## Install

```
npm install cypress --save-dev
yarn add cypress --dev
```

## Run Quick Cypress Test

```
cd test/e2e

docker-compose -f hydra-compose.yml -f php-compose.yml -f cypress-compose.yml  up

```

## Run Cypress on Dev

### Start development server
```
# server must expose port to host machine

$ docker-compose -f hydra-compose.yml -f php-compose.yml up
```

### Start Cypress
```
# server must expose port to host machine

$ export CYPRESS_BASE_URL=http://10.254.254.254:8000; $(npm bin)/cypress open
```


## Writing Tests with Cypress
### Cypress Folder Structure
```
# https://docs.cypress.io/guides/core-concepts/writing-and-organizing-tests.html#Folder-Structure

cypress.json
cypress
    |---- fixtures - JSON files of common data objects needed in tests
    |---- integration - all of our tests, we would often create sub-folders per page/feature or even by groups of tests
    |---- pages - page objects and each feature would often have its own sub-folder for the pages related to it
    |---- plugins - custom plugins to run in a Node server, each feature/page would have its own sub-folder for API teardown/setup
    |---- support - custom commands and types here
    |---- utils - extra utility files to be used throughout
    |---- /config - environment configuration JSON files to extend/override the base cypress.json file - not all teams did this but it's another approach

```


## Cypress References
- https://docs.cypress.io/guides/guides/command-line.html#How-to-run-commands
