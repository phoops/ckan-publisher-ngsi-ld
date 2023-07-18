package usecase

import (
	"context"
	"regexp"
	"time"

	entities "bitbucket.org/phoops/odala-mt-earthquake/internal/core/entities"
	ngsild "bitbucket.org/phoops/odala-mt-earthquake/internal/infrastructure/ngsi-ld"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DataFetcher interface {
	FetchData(ctx context.Context, beginData time.Time, offset int) (entities.Vehicles, error)
}

type DataPersistor interface {
	GetLastUpdate(ctx context.Context) (time.Time, error)
	WriteData(ctx context.Context, data []entities.GateCount) error
}

type FetchAndPush struct {
	logger            *zap.SugaredLogger
	fetcher           DataFetcher
	persistor         DataPersistor
	aggregateInterval int
}

func NewFetchAndPush(
	logger *zap.SugaredLogger,
	fetcher *ngsild.Client,
	persistor DataPersistor,
	aggregateInterval int,
) (*FetchAndPush, error) {
	if logger == nil || fetcher == nil || persistor == nil || aggregateInterval <= 0 {
		return nil, errors.New("all parameters must be non-nil")
	}
	logger = logger.With("usecase", "FetchAndPush")

	return &FetchAndPush{
		logger,
		fetcher,
		persistor,
		aggregateInterval,
	}, nil
}

func (fp *FetchAndPush) Execute(ctx context.Context) error {

	// get last update from CKAN
	lastUpdate, err := fp.persistor.GetLastUpdate(ctx)
	fp.logger.Infow("last update", "date", lastUpdate)
	if err != nil {
		fp.logger.Errorw("can't get last update", "error", err)
		return errors.Wrap(err, "can't get last update")
	}
	beginDate := lastUpdate.Add(1 * time.Second)
	if beginDate.After(time.Now().Add(- time.Duration(fp.aggregateInterval) * time.Minute)) {
		fp.logger.Infow("no new data", "begin date", beginDate)
		return nil
	}


	// fetch data from broker
	offset := 0	// used because the API returns only 1000 records at a time
	countMap := make(map[string]map[time.Time]entities.GateCount)

	for {
		fetchedData, err := fp.fetcher.FetchData(ctx, beginDate, offset)
		if err != nil {
			fp.logger.Errorw("can't fetch data", err)
			return errors.Wrap(err, "can't fetch data")
		}
		fp.logger.Infow("fetched data", "begin date", beginDate, "count", len(fetchedData), "offset", offset)

		countMap, err = fp.Aggregate(fetchedData, countMap)
		if err != nil {
			fp.logger.Errorw("can't aggregate and convert data", "error", err)
			return errors.Wrap(err, "can't aggregate and convert data")
		}

		if len(fetchedData) < 1000 {
			break
		}
		offset += 1000
	}

	// convert map to slice
	var vechicleRecords []entities.GateCount
	for _, innerMap := range countMap {
		for _, gateCount := range innerMap {
			if gateCount.BeginObservation.Before(time.Now().Add(- time.Duration(fp.aggregateInterval) * time.Minute)) {
				vechicleRecords = append(vechicleRecords, gateCount)
			}
		}
	}

	// write data
	err = fp.persistor.WriteData(ctx, vechicleRecords)
	if err != nil {
		fp.logger.Errorw("can't write data", "error", err)
		return errors.Wrap(err, "can't write data")
	}
	fp.logger.Infow("aggregate data written", "count", len(vechicleRecords))

	return nil
}

// Aggregate vehicles data by parking and gate and convert it to GateCount objects that will be stored in CKAN
func (fp *FetchAndPush) Aggregate(vechicles entities.Vehicles, countMap map[string]map[time.Time]entities.GateCount) (map[string]map[time.Time]entities.GateCount, error) {

	re := regexp.MustCompile(`Parking: (\S+), Gate: (\S+)`)

	for _, v := range vechicles {

		beginDate := v.Location.ObservedAt.Truncate(time.Duration(fp.aggregateInterval) * time.Minute)

		matches := re.FindStringSubmatch(v.Description.Value)
		if len(matches) != 3 {
			fp.logger.Errorw("can't parse description. Skipped", "description", v.Description.Value)
			continue
		}
		parking := matches[1]
		gate := matches[2]

		if countMap[v.Description.Value] == nil {
			countMap[v.Description.Value] = make(map[time.Time]entities.GateCount)
		}

		gateCount, exists := countMap[v.Description.Value][beginDate]
		if !exists {
			gateCount = entities.GateCount{
				Parking:          parking,
				Gate:             gate,
				Coordinate1:      v.Location.Value.Coordinates[1], //they are inverted because of the mt problem (check readme)
				Coordinate2:      v.Location.Value.Coordinates[0],
				BeginObservation: beginDate,
				EndObservation:   beginDate.Add(time.Duration(fp.aggregateInterval - 1)*time.Minute + 59*time.Second),
				Count:            0,
			}
		}
		gateCount.Count++
		countMap[v.Description.Value][beginDate] = gateCount
	}


	return countMap, nil
}

