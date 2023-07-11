package persistor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	entities "bitbucket.org/phoops/odala-mt-earthquake/internal/core/entities"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Client struct {
	logger    *zap.SugaredLogger
	baseURL   string
	dataStore string
	key       string
	interval  int
}

func NewClient(logger *zap.SugaredLogger, baseURL string, dataStore string, key string, interval int) (*Client, error) {
	if logger == nil || baseURL == "" || dataStore == "" || key == "" {
		return nil, errors.New("all parameters must be non-nil")
	}
	logger = logger.With("component", "persistor client")

	return &Client{
		logger,
		baseURL,
		dataStore,
		key,
		interval,
	}, nil
}

// get last record date from CKAN. If no record is found, return one year ago
func (c *Client) GetLastUpdate(ctx context.Context) (time.Time, error) {

	reqBody := entities.ReadRecordBody{
		ResourceId: c.dataStore,
		Limit:      1,
		Sort:       "endObservation desc",
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Errorw("can't marshal request body", "err", err)
		return time.Now(), errors.Wrap(err, "can't marshal request body")
	}

	url := c.baseURL + "/api/3/action/datastore_search"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.key)
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Errorw("can't read data", "err", err)
		return time.Now(), errors.Wrap(err, "can't  write data")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.logger.Fatal("error reading data from CKAN. Code: ", resp.StatusCode)
		return time.Now(), errors.New("error reading data from CKAN")
	}
	
	var data map[string]interface{}
	respBodyBytes, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal([]byte(respBodyBytes), &data)
	if err != nil {
		c.logger.Fatalw("can't unmarshal response body", "err", err)
		return time.Now(), errors.Wrap(err, "can't unmarshal response body")
	}

	records := data["result"].(map[string]interface{})["records"].([]interface{})
	if len(records) == 0 {
		c.logger.Info(fmt.Sprintf("no record found in CKAN. Beginning from %d minutes ago", c.interval))
		return time.Now().Add( - time.Duration(-c.interval * 2) * time.Minute), nil
	}
	record := records[0].(map[string]interface{})
	bucketStartTimestamp := record["endObservation"].(string)
	t, err := time.Parse("2006-01-02T15:04:05", bucketStartTimestamp)
	if err != nil {
		c.logger.Errorw("can't parse endObservation", "err", err)
		return time.Now(), errors.Wrap(err, "can't parse endObservation")
	}

	c.logger.Infow("updating from last date", "last date found in CKAN", t)

	return t, nil
}

// write data to CKAN
func (c *Client) WriteData(ctx context.Context, data []entities.GateCount) error {

	reqBody := entities.WriteRequestBody{
		Records:    data,
		ResourceId: c.dataStore,
		Force:      "true",
		Method:     "insert",
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Errorw("can't marshal request body", "err", err)
		return errors.Wrap(err, "can't marshal request body")
	}

	url := c.baseURL + "/api/3/action/datastore_upsert"

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.key)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Errorw("can't write data", "err", err)
		return errors.Wrap(err, "can't  write data")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorw("can't write data. Code: ", resp.StatusCode)
		return errors.Wrap(err, "can't  write data")
	}

	return nil
}
