package main

import (
	"context"

	"bitbucket.org/phoops/odala-mt-earthquake/internal/core/usecase"
	"bitbucket.org/phoops/odala-mt-earthquake/internal/infrastructure/persistor"
	"bitbucket.org/phoops/odala-mt-earthquake/internal/infrastructure/config"
	ngsild "bitbucket.org/phoops/odala-mt-earthquake/internal/infrastructure/ngsi-ld"
	"github.com/pkg/errors"
	"go.uber.org/zap"

)

func main() {

	// Logger
	sourLogger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	logger := sourLogger.Sugar()

	// Configuration
	conf, err := config.LoadEarthquakeConfig()
	if err != nil {
		errMsg := errors.Wrap(err, "cannot read configuration").Error()
		logger.Fatal(errMsg)
	}

	// Context Broker Client
	contextBrokerClient, err := ngsild.NewClient(
		logger,
		conf.BrokerURL,
	)
	if err != nil {
		errMsg := errors.Wrap(err, "cannot instantiate context broker client").Error()
		logger.Fatal(errMsg)
	}

	// CKAN Persistor
	persistor, err := persistor.NewClient(logger, conf.CkanURL, conf.CkanDatastore, conf.CkanKey)
	if err != nil {
		errMsg := errors.Wrap(err, "cannot create data persistor").Error()
		logger.Fatal(errMsg)
	}

	// Usecase
	fetchAndPush, err := usecase.NewFetchAndPush(logger, contextBrokerClient, persistor)
	if err != nil {
		errMsg := errors.Wrap(err, "cannot create usecase").Error()
		logger.Fatal(errMsg)
	}

	// Execute
	err = fetchAndPush.Execute(
		context.Background(),
	)
	if err != nil {
		errMsg := errors.Wrap(err, "cannot sync vehicles on context broker").Error()
		logger.Fatal(errMsg)
	}
}
