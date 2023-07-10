# Earthquake

## Description
A job to take Vehicle data from Scorpio Broker, aggregates it and sends it to Scorpio Broker.

NOTE 1: Coordinates are swapped in the Scorpio Broker, so I swapped them back before sending them to the CKAN
NOTE 2: Scorpio Broker has a limit to 1000 element per request. Maybe there is a workaround for this.

## Before Running First Time


## Run Dev
`go run cmd/nurse/main.go`