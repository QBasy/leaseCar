package adapters

import (
    "context"

    meili "github.com/meilisearch/meilisearch-go"
)

type MeiliAdapter struct {
    client *meili.Client
}

func NewMeiliAdapter(c *meili.Client) *MeiliAdapter {
    return &MeiliAdapter{client: c}
}

func (m *MeiliAdapter) IndexLease(ctx context.Context, index string, doc interface{}) error {
    _, err := m.client.Index(index).AddDocuments(doc)
    return err
}

func (m *MeiliAdapter) Search(ctx context.Context, index, q string, limit int) (*meili.SearchResponse, error) {
    res, err := m.client.Index(index).Search(q, &meili.SearchRequest{Limit: int64(limit)})
    return res, err
}
