# Luxor Interview

## Overview

This mock Stratum server implements the `mining.authorize` and `mining.subscribe` client methods and includes a ginkgo test to test them. Manual testing was done with telnet.

## How to run

First start the database with:

`docker-compose up`

Then

`go run cmd/luxor/main.go`

## How to test

First start the database with:

`docker-compose up`

Then

`ginkgo -v`
