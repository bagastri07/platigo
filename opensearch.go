package platigo

import (
	"context"
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/goccy/go-json"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
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

type OpensearchClient interface {
	// Index indexes a document in OpenSearch.
	Index(ctx context.Context, indexName string, model IndexModel) (*opensearchapi.Response, error)

	// CreateIndices creates an index in OpenSearch.
	CreateIndices(ctx context.Context, indexName string, body *strings.Reader) (*opensearchapi.Response, error)

	// PutIndicesMapping updates the mapping for one or more indices in OpenSearch.
	PutIndicesMapping(ctx context.Context, indexNames []string, body *strings.Reader) (*opensearchapi.Response, error)

	// Search performs a search query in OpenSearch.
	Search(ctx context.Context, indexNames []string, body *strings.Reader) (*opensearchapi.Response, error)
}

type opensearchClient struct {
	client *opensearch.Client
}

// NewOpensearchClient creates a new OpensearchClient instance.
func NewOpensearchClient(config *OSConfig) (OpensearchClient, error) {
	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify}, // #nosec G402
		},
		Addresses: config.Addresses,
		Username:  config.Username,
		Password:  config.Password,
	})
	platigoOSClient := &opensearchClient{
		client: client,
	}

	return platigoOSClient, err
}

func (k *opensearchClient) CreateIndices(ctx context.Context, indexName string, body *strings.Reader) (*opensearchapi.Response, error) {
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

func (k *opensearchClient) PutIndicesMapping(ctx context.Context, indexNames []string, body *strings.Reader) (*opensearchapi.Response, error) {
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

func (k *opensearchClient) Index(ctx context.Context, indexName string, model IndexModel) (*opensearchapi.Response, error) {
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

func (k *opensearchClient) Search(ctx context.Context, indexNames []string, body *strings.Reader) (*opensearchapi.Response, error) {
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
