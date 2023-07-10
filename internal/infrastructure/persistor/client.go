package persistor

import (
	"bytes"
	"context"
	"encoding/json"
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
}

func NewClient(logger *zap.SugaredLogger, baseURL string, dataStore string, key string) (*Client, error) {
	if logger == nil || baseURL == "" || dataStore == "" || key == "" {
		return nil, errors.New("all parameters must be non-nil")
	}
	logger = logger.With("component", "persistor client")

	return &Client{
		logger,
		baseURL,
		dataStore,
		key,
	}, nil
}

// get last record date from CKAN. If no record is found, return one year ago
func (c *Client) GetLastUpdate(ctx context.Context) (time.Time, error) {

	reqBody := entities.ReadRecordBody{
		ResourceId: c.dataStore,
		Limit:      1,
		Sort:       "observedAt",
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

	respBody := resp.Body
	bodyBytes, _ = ioutil.ReadAll(respBody)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(bodyBytes), &data)
	if err != nil {
		c.logger.Fatalw("can't unmarshal response body", "err", err)
		return time.Now(), errors.Wrap(err, "can't unmarshal response body")
	}

	records := data["result"].(map[string]interface{})["records"].([]interface{})
	if len(records) == 0 {
		c.logger.Infow("no records found. Begin from one year ago")
		return time.Now().AddDate(0, 0, -1), nil //TODO change to one year?
	}
	record := records[0].(map[string]interface{})
	bucketStartTimestamp := record["observedAt"].(string)

	t, err := time.Parse("2006-01-02T15:04:05", bucketStartTimestamp)

	if err != nil {
		c.logger.Errorw("can't parse observedAt", "err", err)
		return time.Now(), errors.Wrap(err, "can't parse observedAt")
	}

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
		c.logger.Errorw("can't write data", "err", err)
		return errors.Wrap(err, "can't  write data")
	}

	return nil
}
