package ngsild

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"net/url"
	"io/ioutil"
	"time"

	entities "bitbucket.org/phoops/odala-mt-earthquake/internal/core/entities"
	"github.com/phoops/ngsi-gold/client"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Client struct {
	logger       *zap.SugaredLogger
	baseURL      string
	ngsiLdClient *client.NgsiLdClient
}

func NewClient(logger *zap.SugaredLogger, baseURL string) (*Client, error) {
	if logger == nil {
		return nil, errors.New("all parameters must be non-nil")
	}
	logger = logger.With("component", "NGSI-LD client")
	ngsiLdClient, err := client.New(
		client.SetURL(baseURL),
	)
	if err != nil {
		return nil, errors.Wrap(err, "can't instantiate ngsi-ld client")
	}

	return &Client{
		logger,
		baseURL,
		ngsiLdClient,
	}, nil
}


// Reads vehicles on the broker observed after beginDate
func (c *Client) FetchData(ctx context.Context, beginDate time.Time, offset int) (entities.Vehicles, error) {

	// create request
	encodedDatetime := url.QueryEscape(beginDate.Format("2006-01-02T15:04:05Z"))
	queryParams := url.Values{}
	queryParams.Add("type", "Vehicle")
	queryParams.Add("limit", "1000")
	queryParams.Add("q", fmt.Sprintf("location.observedAt>=%s", encodedDatetime))
	queryParams.Add("offset", fmt.Sprintf("%d", offset))

	url := fmt.Sprintf("%s//ngsi-ld/v1/entities/?%s", c.baseURL, queryParams.Encode())
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c.logger.Errorw("can't create request", "err", err)
		return nil, errors.Wrap(err, "can't create request")
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Link", `<https://raw.githubusercontent.com/smart-data-models/data-models/master/context.jsonld>; rel="http://www.w3.org/ns/json-ld#context"; type="application/ld+json"`)
	client := http.DefaultClient

	// send request
	resp, err := client.Do(req)
	if err != nil {
		c.logger.Errorw("can't read vehicles from broker", "err", err)
		return nil, errors.Wrap(err, "can't read vehicles from broker")
	}
	defer resp.Body.Close()

	// convert response
	vehiclesBody, _ := ioutil.ReadAll(resp.Body)
	vehicleResponse := entities.Vehicles{}
	err = json.Unmarshal(vehiclesBody, &vehicleResponse)
	if err != nil {
		c.logger.Errorw("error decoding JSON:", err)
		return nil, errors.Wrap(err, "error decoding JSON")
	}

	return vehicleResponse, nil

}
