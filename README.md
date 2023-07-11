# Earthquake

## Description
A job to take Vehicle data from Scorpio Broker, aggregates it and sends it to Scorpio Broker.

NOTE 1: Coordinates are swapped in the Scorpio Broker, so I swapped them back before sending them to the CKAN
NOTE 2: Scorpio Broker has a limit to 1000 element per request. Maybe there is a workaround for this.

## Before Running First Time

Make sure you have a dataset on CKAN with a resource (datastore). If you don't have one:

1. Create a dataset on CKAN using the web interface (regione toscana's CKAN has problems with the API)
2. Create a resource (datastore) on CKAN using the API. Curl:

```
curl --location 'https://dati.toscana.it/api/3/action/datastore_create' \
--header 'Authorization: 5fea970e-9fe2-4323-9ad7-66b4f1be3839' \
--header 'Content-Type: application/json' \
--data '{ 
    "resource": {
        "package_id": "passaggio-veicoli-ai-parcheggi",
        "name": "Passaggi alle sbarre"
                }, 
    "fields": [
        {"id": "parking", "type": "text"}, 
        {"id": "gate", "type": "text"},
        {"id": "beginObservation", "type": "timestamp"},
        {"id": "endObservation", "type": "timestamp"},
        {"id": "coordinate1", "type": "float8"},
        {"id": "coordinate2", "type": "float8"},
        {"id": "count", "type": "int"}
        ]
    }'
```
(remember to change the authorization token and the package_id)

## Run Dev
`go run cmd/nurse/main.go`