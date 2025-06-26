package client

import (
	"context"

	"github.com/elastic/go-elasticsearch/v9/esutil"
)

type (
	bulkIndexerImpl struct {
		es esutil.BulkIndexer
	}
)

func newBulkProcessor_n(cfg esutil.BulkIndexerConfig) (*bulkIndexerImpl, error) {
	indexer, err := esutil.NewBulkIndexer(cfg)
	if err != nil {
		return nil, err
	}
	return &bulkIndexerImpl{
		es: indexer,
	}, nil
}

func (b *bulkIndexerImpl) Add(ctx context.Context, request *BulkIndexerRequest) error {
	switch request.RequestType {
	case BulkableRequestTypeIndex:
		bulkIndexRequest := esutil.BulkIndexerItem{
			Index:       request.Index,
			Action:      "index",
			DocumentID:  request.ID,
			Version:     request.Version,
			VersionType: versionTypeExternal,
			Body:        request.Doc,
		}
		return b.es.Add(ctx, bulkIndexRequest)

	case BulkableRequestTypeDelete:
		bulkDeleteRequest := esutil.BulkIndexerItem{
			Index:       request.Index,
			Action:      "delete",
			DocumentID:  request.ID,
			Version:     request.Version,
			VersionType: versionTypeExternal,
		}
		return b.es.Add(ctx, bulkDeleteRequest)
	}
	return nil
}

func (b *bulkIndexerImpl) Close(ctx context.Context) error {
	return b.es.Close(ctx)
}
