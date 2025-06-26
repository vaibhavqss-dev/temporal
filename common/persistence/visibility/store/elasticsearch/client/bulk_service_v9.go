package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v9"
)

type (
	bulkServiceImpl_n struct {
		client *elasticsearch.Client
		buf    []string
		mu     sync.Mutex
	}
)

func newBulkService_n(client *elasticsearch.Client) *bulkServiceImpl_n {
	return &bulkServiceImpl_n{
		client: client,
		buf:    make([]string, 0),
	}
}

func (b *bulkServiceImpl_n) Do(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.buf) == 0 {
		return nil
	}

	body := strings.Join(b.buf, "\n") + "\n"
	res, err := b.client.Bulk(
		strings.NewReader(body),
		b.client.Bulk.WithContext(ctx),
	)

	b.buf = b.buf[:0]

	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk request failed: %s", res.String())
	}

	return nil
}

func (b *bulkServiceImpl_n) NumberOfActions() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.buf)
}

func (b *bulkServiceImpl_n) Add(request *BulkIndexerRequest) {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch request.RequestType {
	case BulkableRequestTypeIndex:
		meta := map[string]map[string]interface{}{
			"index": {
				"_index":       request.Index,
				"_id":          request.ID,
				"version_type": versionTypeExternal,
				"version":      request.Version,
			},
		}
		metaJSON, _ := json.Marshal(meta)	
		docJSON, _ := json.Marshal(request.Doc)

		b.buf = append(b.buf, string(metaJSON))
		b.buf = append(b.buf, string(docJSON))

	case BulkableRequestTypeDelete:
		meta := map[string]map[string]interface{}{
			"delete": {
				"_index":       request.Index,
				"_id":          request.ID,
				"version_type": versionTypeExternal,
				"version":      request.Version,
			},
		}
		metaJSON, _ := json.Marshal(meta)
		b.buf = append(b.buf, string(metaJSON))
	}
}
