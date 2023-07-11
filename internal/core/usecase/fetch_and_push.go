package usecase

import (
	"context"
	"time"
	"regexp"

	entities "bitbucket.org/phoops/odala-mt-earthquake/internal/core/entities"
	ngsild "bitbucket.org/phoops/odala-mt-earthquake/internal/infrastructure/ngsi-ld"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DataFetcher interface {
	FetchData(ctx context.Context, beginData time.Time) (entities.Vehicles, error)
}

type DataPersistor interface {
	GetLastUpdate(ctx context.Context) (time.Time, error)
	WriteData(ctx context.Context, data []entities.GateCount) error
}

type FetchAndPush struct {
	logger    *zap.SugaredLogger
	fetcher   DataFetcher
	persistor DataPersistor
}

func NewFetchAndPush(
	logger *zap.SugaredLogger,
	fetcher *ngsild.Client,
	persistor DataPersistor,
) (*FetchAndPush, error) {
	if logger == nil || fetcher == nil || persistor == nil {
		return nil, errors.New("all parameters must be non-nil")
	}
	logger = logger.With("usecase", "FetchAndPush")

	return &FetchAndPush{
		logger,
		fetcher,
		persistor,
	}, nil
}

func (fp *FetchAndPush) Execute(ctx context.Context) error {

	lastUpdate, err := fp.persistor.GetLastUpdate(ctx)
	if err != nil {
		fp.logger.Errorw("can't get last update", "error", err)
		return errors.Wrap(err, "can't get last update")
	}
	beginDate := lastUpdate.Add(1 * time.Second)

	if beginDate.After(time.Now().Add(15*time.Minute)) {
		fp.logger.Infow("no new data", "begin date", beginDate)
		return nil
	}



	fetchedData, err := fp.fetcher.FetchData(ctx, beginDate)
	if err != nil{
		fp.logger.Errorw("can't fetch data", err)
		return errors.Wrap(err, "can't fetch data")
	}
	fp.logger.Infow("fetched data", "count", len(fetchedData), "begin date", beginDate)

	vechicleRecords, err := fp.AggregateAndConvert(fetchedData)
	if err != nil {
		fp.logger.Errorw("can't aggregate and convert data", "error", err)
		return errors.Wrap(err, "can't aggregate and convert data")
	}

	err = fp.persistor.WriteData(ctx, vechicleRecords)
	if err != nil {
		fp.logger.Errorw("can't write data", "error", err)
		return errors.Wrap(err, "can't write data")
	}
	fp.logger.Infow("agggregate data written", "count", len(vechicleRecords))

	return nil
}

// Aggregate vehicles data by parking and gate and convert it to GateCount objects that will be stored in CKAN
func (fp *FetchAndPush) AggregateAndConvert(vechicles entities.Vehicles) ([]entities.GateCount, error) {

	countMap := make(map[string]map[time.Time]entities.GateCount)
	re := regexp.MustCompile(`Parking: (\S+), Gate: (\S+)`)

	for _, v := range vechicles {

		beginDate := v.Location.ObservedAt.Truncate(15 * time.Minute)
		
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
				Coordinate1:      v.Location.Value.Coordinates[1],   //they are inverted because of the mt problem (check readme)
				Coordinate2:      v.Location.Value.Coordinates[0],
				BeginObservation: beginDate,
				EndObservation:   beginDate.Add(14*time.Minute + 59*time.Second),
				Count:            0,
			}
		}
		gateCount.Count++
		countMap[v.Description.Value][beginDate] = gateCount

	}

	var countColl []entities.GateCount

	for _, innerMap := range countMap {
		for _, gateCount := range innerMap {
			countColl = append(countColl, gateCount)
		}
	}

	return countColl, nil
}