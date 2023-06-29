package platigo

import (
	"context"
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/bagastri07/platigo/utils"
	"github.com/goccy/go-json"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"github.com/sirupsen/logrus"
)

type OSConfig struct {
	Addresses          []string
	InsecureSkipVerify bool // Set to true only if SSL certificate verification is intentionally skipped for specific use cases (e.g., testing or development).
	Username           string
	Password           string
}

type IndexModel interface {
	GetID() string
}

type OpenSearchClient interface {
	// Index indexes a document in OpenSearch.
	Index(ctx context.Context, indexName string, model IndexModel) (*opensearchapi.Response, error)

	// CreateIndices creates an index in OpenSearch.
	CreateIndices(ctx context.Context, indexName string, body *strings.Reader) (*opensearchapi.Response, error)

	// PutIndicesMapping updates the mapping for one or more indices in OpenSearch.
	PutIndicesMapping(ctx context.Context, indexNames []string, body *strings.Reader) (*opensearchapi.Response, error)

	// Search performs a search query in OpenSearch.
	Search(ctx context.Context, indexNames []string, body *strings.Reader) (*opensearchapi.Response, error)

	// BulkIndex indexes multiple documents in OpenSearch.
	BulkIndex(ctx context.Context, indexName string, models []IndexModel) error

	// Ping pings the OpenSearch cluster to check its availability.
	Ping(ctx context.Context) (*opensearchapi.Response, error)
}

type openSearchClient struct {
	client *opensearch.Client
}

// NewOpenSearchClient creates a new OpenSearchClient instance.
func NewOpenSearchClient(config *OSConfig) (OpenSearchClient, error) {
	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify}, // #nosec G402
		},
		Addresses: config.Addresses,
		Username:  config.Username,
		Password:  config.Password,
	})
	platigoOSClient := &openSearchClient{
		client: client,
	}

	return platigoOSClient, err
}

func (k *openSearchClient) CreateIndices(ctx context.Context, indexName string, body *strings.Reader) (*opensearchapi.Response, error) {
	logger := logrus.WithFields(logrus.Fields{
		"indexName": indexName,
	})
	req := opensearchapi.IndicesCreateRequest{
		Index: indexName,
		Body:  body,
	}

	res, err := req.Do(ctx, k.client)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	logger.Info(res)

	return res, err
}

func (k *openSearchClient) PutIndicesMapping(ctx context.Context, indexNames []string, body *strings.Reader) (*opensearchapi.Response, error) {
	logger := logrus.WithFields(logrus.Fields{
		"indexNames": indexNames,
	})

	req := opensearchapi.IndicesPutMappingRequest{
		Index: indexNames,
		Body:  body,
	}

	res, err := req.Do(ctx, k.client)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	logger.Info(res)

	return res, nil

}

func (k *openSearchClient) Index(ctx context.Context, indexName string, model IndexModel) (*opensearchapi.Response, error) {
	logger := logrus.WithFields(logrus.Fields{
		"indexName": indexName,
		"docID":     model.GetID(),
	})

	docData, err := json.Marshal(model)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	body := strings.NewReader(string(docData))

	req := opensearchapi.IndexRequest{
		Index:      indexName,
		DocumentID: model.GetID(),
		Body:       body,
		Pretty:     true,
	}

	res, err := req.Do(ctx, k.client)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	logger.Info(res)

	return res, nil
}

func (k *openSearchClient) Search(ctx context.Context, indexNames []string, body *strings.Reader) (*opensearchapi.Response, error) {
	logger := logrus.WithFields(logrus.Fields{
		"indexNames": indexNames,
	})

	req := opensearchapi.SearchRequest{
		Index:  indexNames,
		Body:   body,
		Pretty: true,
	}

	res, err := req.Do(ctx, k.client)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	logger.Info(res)

	return res, nil
}

func (k *openSearchClient) BulkIndex(ctx context.Context, indexName string, models []IndexModel) error {
	logger := logrus.WithFields(logrus.Fields{
		"indexName": indexName,
	})

	bulkIndexer, err := opensearchutil.NewBulkIndexer(opensearchutil.BulkIndexerConfig{
		Index:      indexName,
		Client:     k.client,
		NumWorkers: 10,
	})

	if err != nil {
		logger.Errorf("Failed to create bulk indexer: %s", err)
		return err
	}

	for _, model := range models {
		docID := model.GetID()
		jsonData, err := json.Marshal(model)
		if err != nil {
			logger.Error(err)
			continue
		}

		item := opensearchutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: docID,
			Body:       strings.NewReader(string(jsonData)),
		}

		err = bulkIndexer.Add(ctx, item)
		if err != nil {
			logger.Errorf("Failed to add document ID %s to bulk indexer: %s", docID, err)
		}
	}

	err = bulkIndexer.Close(ctx)
	if err != nil {
		logger.Errorf("Failed to close bulk indexer: %s", err)
		return err
	}

	stat := bulkIndexer.Stats()
	logger.Info("Bulk Indexer Stat: ", utils.Dump(stat))

	return nil

}

func (k *openSearchClient) Ping(ctx context.Context) (*opensearchapi.Response, error) {
	req := opensearchapi.PingRequest{}

	res, err := req.Do(ctx, k.client)
	if err != nil {
		logrus.WithError(err).Error("Failed to ping OpenSearch cluster")
		return nil, err
	}

	return res, nil
}
