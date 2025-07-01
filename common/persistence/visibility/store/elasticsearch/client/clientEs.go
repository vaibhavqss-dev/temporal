//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination client_mock.go

package client

import (
	"context"
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	enumspb "go.temporal.io/api/enums/v1"
)

const (
	versionTypeExternal                 = "external"
	minimumCloseIdleConnectionsInterval = 15 * time.Second
)

type (
	// NewES_Client is a wrapper around Elasticsearch client library.
	NewEsClient interface {
		Get(ctx context.Context, index string, docID string) (*types.GetResult, error)
		Search(ctx context.Context, p *NewEsSearchParameters) (*search.Response, error)
		Count(ctx context.Context, index string, query types.Query) (int64, error)
		CountGroupBy(ctx context.Context, index string, query types.Query, aggName string, agg map[string]interface{}) (*search.Response, error)
		RunBulkProcessor(ctx context.Context, p *BulkIndexerParameters) (BulkIndexer, error)

		// TODO (alex): move this to some admin client (and join with IntegrationTestsClient)
		PutMapping(ctx context.Context, index string, mapping map[string]enumspb.IndexedValueType) (bool, error)
		WaitForYellowStatus(ctx context.Context, index string) (string, error)
		GetMapping(ctx context.Context, index string) (map[string]string, error)
		IndexExists(ctx context.Context, indexName string) (bool, error)
		CreateIndex(ctx context.Context, index string, body map[string]any) (bool, error)
		DeleteIndex(ctx context.Context, indexName string) (bool, error)
		CatIndices(ctx context.Context, target string) (*[]types.IndicesRecord, error)

		OpenScroll(ctx context.Context, p *NewEsSearchParameters, keepAliveInterval time.Duration) (*search.Response, error)
		Scroll(ctx context.Context, id string, keepAliveInterval time.Duration) (*search.Response, error)
		CloseScroll(ctx context.Context, id string) error

		IsPointInTimeSupported(ctx context.Context) bool
		OpenPointInTime(ctx context.Context, index string, keepAliveInterval time.Duration) (string, error)
		ClosePointInTime(ctx context.Context, id string) (bool, error)
	}

	NewEsCLIClient interface {
		Client
		Delete(ctx context.Context, indexName string, docID string, version int64) error
	}

	NewEsIntegrationTestsClient interface {
		Client
		IndexPutTemplate(ctx context.Context, templateName string, bodyString string) (bool, error)
		IndexPutSettings(ctx context.Context, indexName string, bodyString string) (bool, error)
		IndexGetSettings(ctx context.Context, indexName string) (map[string]*IndicesGetSettingsResponse, error)
		Ping(ctx context.Context) error
	}

	NewEsSearchParameters struct {
		Index       string
		Query       types.Query
		PageSize    int
		Sorter      []types.Sort
		SearchAfter []interface{}
		ScrollID    string
		PointInTime *types.PointInTimeReference
	}
)

type (
	IndicesGetSettingsResponse struct {
		Settings map[string]interface{} `json:"settings"`
	}
)
