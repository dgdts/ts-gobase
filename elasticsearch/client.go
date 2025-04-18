package elasticsearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
)

var (
	gClient *esClient
	once    sync.Once
)

type esClient struct {
	es *elasticsearch.Client
}

func InitEsClient(cfg Config) error {
	var err error
	once.Do(func() {
		var client *esClient
		client, err = newClient(cfg)
		if err != nil {
			return
		}
		gClient = client
	})
	return err
}

type Config struct {
	Addresses []string `yaml:"addresses"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
}

func newClient(cfg Config) (*esClient, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,

		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// suggest provide real certificate when using production ES.
				InsecureSkipVerify: true,
			},
		},
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	return &esClient{es: client}, nil
}

// Index indexes a document in Elasticsearch.
func Index(ctx context.Context, index string, id string, doc any) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	_, err = gClient.es.Index(
		index,
		bytes.NewReader(body),
		gClient.es.Index.WithContext(ctx),
		gClient.es.Index.WithDocumentID(id),
	)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}

	return nil
}

// Get a document from Elasticsearch with the given index and ID.
func Get(ctx context.Context, index string, id string, out any) error {
	res, err := gClient.es.Get(
		index,
		id,
		gClient.es.Get.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}
	defer res.Body.Close()

	var response struct {
		Source json.RawMessage `json:"_source"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if err := json.Unmarshal(response.Source, out); err != nil {
		return fmt.Errorf("failed to unmarshal document: %w", err)
	}

	return nil
}

// Delete
func Delete(ctx context.Context, index string, id string) error {
	_, err := gClient.es.Delete(
		index,
		id,
		gClient.es.Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// Search
func Search(ctx context.Context, index string, query map[string]any, out any) error {
	body, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	res, err := gClient.es.Search(
		gClient.es.Search.WithContext(ctx),
		gClient.es.Search.WithIndex(index),
		gClient.es.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return fmt.Errorf("failed to perform search: %w", err)
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func GetRawClient() *elasticsearch.Client {
	return gClient.es
}
