package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esapi"
	"go.temporal.io/server/common/log"
)

type (
	ESClient struct {
		ESClient                   *elasticsearch.Client
		url                        url.URL
		initIsPointInTimeSupported sync.Once
		isPointInTimeSupported     bool
	}
)

func NewESClient(cfg *Config, httpClient *http.Client, logger log.Logger) (*ESClient, error) {
	var urls []string
	if len(cfg.URLs) > 0 {
		urls = make([]string, len(cfg.URLs))
		for i, u := range cfg.URLs {
			urls[i] = u.String()
		}
	} else {
		urls = []string{cfg.URL.String()}
	}

	esCfg := elasticsearch.Config{
		Addresses:           urls,
		Username:            cfg.Username,
		Password:            cfg.Password,
		CompressRequestBody: true,
		Transport:           cfg.httpClient.Transport,
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, err
	}

	return &ESClient{
		ESClient: client,
		url:      cfg.URL,
	}, nil
}

func (c *ESClient) Get(ctx context.Context, index string, docID string) (*esapi.Response, error) {
	res, err := c.ESClient.Get(index, docID)
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("error getting document %s: %s", docID, res.String())
	}
	return res, nil
}

func (c *ESClient) Search(ctx context.Context, p *SearchParameters) (*esapi.Response, error) {
	query := map[string]interface{}{
		"query":            p.Query,
		"sort":             p.Sorter,
		"track_total_hits": false,
	}

	if p.PageSize != 0 {
		query["size"] = p.PageSize
	}
	if len(p.SearchAfter) > 0 {
		query["search_after"] = p.SearchAfter
	}
	if p.PointInTime != nil {
		query["point_in_time"] = map[string]interface{}{
			"id":         p.PointInTime,
			"keep_alive": p.PointInTime.KeepAlive,
		}
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding search query: %w", err)
	}

	req := c.ESClient.Search
	opts := []func(*esapi.SearchRequest){
		req.WithContext(ctx),
		req.WithBody(&buf),
		req.WithTrackTotalHits(false),
	}

	if p.PointInTime == nil {
		opts = append(opts, req.WithIndex(p.Index))
	}

	res, err := req(opts...)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES search error: %s", res.String())
	}

	return res, nil
}

// TODO
func (c *ESClient) Count(ctx context.Context, index string, query string) (int64, error) {
	req := esapi.CountRequest{
		Index: []string{index},
		Body:  strings.NewReader(query),
	}
	res, err := req.Do(ctx, c.ESClient)
	if err != nil {
		return 0, err
	}
	if res.IsError() {
		return 0, fmt.Errorf("error counting documents in index %s: %s", index, res.String())
	}
	var r struct {
		Count int64 `json:"count"`
	}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return 0, err
	}

	return r.Count, nil
}

// // CountGroupBy counts documents in an index with a specific aggregation.
func (c *ESClient) CountGroupBy(
	ctx context.Context,
	index string,
	query map[string]interface{}, // replaces elastic.Query
	aggName string,
	agg map[string]interface{}, // replaces elastic.Aggregation
) (map[string]interface{}, error) {
	searchBody := map[string]interface{}{
		"query":            query,
		"size":             0,
		"track_total_hits": false,
		"aggs": map[string]interface{}{
			aggName: agg,
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(searchBody); err != nil {
		return nil, fmt.Errorf("error encoding search body: %w", err)
	}

	// Send the search request
	res, err := c.ESClient.Search(
		c.ESClient.Search.WithContext(ctx),
		c.ESClient.Search.WithIndex(index),
		c.ESClient.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	aggs, ok := result["aggregations"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing 'aggregations' in response")
	}
	return aggs, nil
}

func (c *ESClient) GetMapping(ctx context.Context, index string) (map[string]string, error) {
	req := esapi.IndicesGetMappingRequest{
		Index: []string{index},
	}
	res, err := req.Do(ctx, c.ESClient)
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("error getting mapping for index %s: %s", index, res.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return convertMappingBody(body, index), nil
}

// // NotImplemented methods for BulkProcessor
// // TODO: IMPLEMENT BULK PROCESSOR
func (c *ESClient) RunBulkProcessor(ctx context.Context, p *BulkProcessorParameters) (BulkProcessor, error) {
	return nil, nil
}

// // TODO
func (c *ESClient) Delete(ctx context.Context, index string, docID string, version int64) error {
	versions := int(version)
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: docID,
		Version:    &versions,
	}

	res, err := req.Do(ctx, c.ESClient)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("error deleting document %s: %s", docID, res.String())
	}
	return nil
}

func (c *ESClient) DeleteIndex(ctx context.Context, indexName string) (bool, error) {
	req := esapi.IndicesDeleteRequest{
		Index: []string{indexName},
	}

	res, err := req.Do(ctx, c.ESClient)
	if err != nil {
		return false, err
	}
	if res.IsError() {
		return false, fmt.Errorf("error deleting index %s: %s", indexName, res.String())
	}
	return true, nil
}

func (c *ESClient) CreateIndex(ctx context.Context, index string, body map[string]any) (bool, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return false, fmt.Errorf("error encoding index body: %w", err)
	}

	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  &buf,
	}
	res, err := req.Do(ctx, c.ESClient)
	if err != nil {
		return false, err
	}
	if res.IsError() {
		return false, fmt.Errorf("error creating index %s: %s", index, res.String())
	}
	return true, nil
}

func (c *ESClient) IndexExists(ctx context.Context, indexName string) (bool, error) {
	req := esapi.IndicesExistsRequest{
		Index: []string{indexName},
	}
	res, err := req.Do(ctx, c.ESClient)
	if err != nil {
		return false, err
	}
	if res.IsError() {
		return false, fmt.Errorf("error checking if index %s exists: %s", indexName, res.String())
	}
	return true, nil
}

func (c *ESClient) CatIndices(ctx context.Context, target string) (*esapi.Response, error) {
	req := esapi.CatIndicesRequest{
		Index: []string{target},
	}
	res, err := req.Do(ctx, c.ESClient)
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("error getting cat indices for target %s: %s", target, res.String())
	}

	// Convert into esapi.Response
	var bytes bytes.Buffer
	if err = json.NewDecoder(res.Body).Decode(&bytes); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *ESClient) IsNotFoundError(err error) bool {
	// return 
}

// func (c *ESClient) PutMapping(ctx context.Context, index string, mapping map[string]enumspb.IndexedValueType) (*esapi.Response, error) {
// 	body := strings.NewReader(fmt.Sprintf(`{"properties": %s}`, mapping))

// 	req := esapi.IndicesPutMappingRequest{
// 		Index: []string{index},
// 		Body:  body,
// 	}

// 	res, err := req.Do(ctx, c.Client)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if res.IsError() {
// 		return nil, fmt.Errorf("error putting mapping for index %s: %s", index, res.String())
// 	}
// 	return res, nil
// }

// func (c *ESClient) WaitForYellowStatus(ctx context.Context, index string) (*esapi.Response, error) {
// 	req := esapi.ClusterHealthRequest{
// 		Index:         []string{index},
// 		WaitForStatus: "yellow",
// 	}
// 	res, err := req.Do(ctx, c.Client)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return res, nil
// }

// func (c *ESClient) OpenPointInTime(ctx context.Context, index string, keepAliveInterval string) (string, error) {
// 	req := esapi.OpenPointInTimeRequest{
// 		Index:     []string{index},
// 		KeepAlive: keepAliveInterval,
// 	}

// 	res, err := req.Do(ctx, c.Client)
// 	if err != nil {
// 		return "", err
// 	}
// 	if res.IsError() {
// 		return "", fmt.Errorf("error opening point in time for index %s: %s", index, res.String())
// 	}

// 	bodyBytes, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("error reading response body: %v", err)
// 	}
// 	return string(bodyBytes), nil
// }

// func (c *ESClient) ClosePointInTime(ctx context.Context, id string) (*esapi.Response, error) {
// 	req := esapi.ClosePointInTimeRequest{
// 		Body: strings.NewReader(fmt.Sprintf(`{"id": "%s"}`, id)),
// 	}

// 	res, err := req.Do(ctx, c.Client)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if res.IsError() {
// 		return nil, fmt.Errorf("error closing point in time %s: %s", id, res.String())
// 	}
// 	return res, nil
// }

// func (c *ESClient) Ping() (*esapi.Response, error) {
// 	res, err := c.Client.Ping()
// 	if err != nil {
// 		return nil, err
// 	}
// 	if res.IsError() {
// 		return nil, fmt.Errorf("error pinging Elasticsearch: %s", res.String())
// 	}
// 	return res, nil
// }
