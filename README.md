
# Platigo

Platigo is a Go utility package that provides functionality for connecting to OpenSearch, an open-source search and analytics engine. It offers a convenient way to interact with OpenSearch clusters and perform various operations such as indexing documents, creating indices, updating mappings, and executing search queries.

## Features

- **OpenSearch Integration**: Platigo enables seamless integration with OpenSearch clusters, allowing you to establish connections and interact with the search engine programmatically.
- **Indexing Documents**: You can easily index documents into the OpenSearch index using the provided utility functions.
- **Creating Indices**: Platigo offers functionality to create indices in OpenSearch, helping you manage the structure of your data.
- **Updating Mappings**: With Platigo, you can update the mapping of one or more indices in OpenSearch, ensuring proper data representation and search functionality.
- **Executing Search Queries**: Platigo allows you to perform search queries on the OpenSearch index, helping you retrieve relevant documents based on specific criteria.

## Installation

To use Platigo in your Go project, you can add it as a dependency using the following command:

```shell
go get github.com/bagastri07/platigo
```

## Usage

To get started with Platigo, create an instance of the OpenSearch client by providing the necessary configuration:

```go
config := &platigo.OSConfig{
    Addresses: []string{"http://opensearch-host:9200"},
    InsecureSkipVerify: false, // Set to true only for specific use cases where SSL certificate verification is intentionally skipped
    Username: "your-username",
    Password: "your-password",
}

client, err := platigo.NewOpensearchClient(config)
if err != nil {
    log.Fatal("Failed to create OpenSearch client:", err)
}
```

Once you have the OpenSearch client, you can use it to perform various operations. Here are a few examples:

**Indexing a Document**

```go
type MyDocument struct {
    ID      string `json:"id"`
    Title   string `json:"title"`
    Content string `json:"content"`
}

doc := &MyDocument{
    ID:      "document-id",
    Title:   "Sample Document",
    Content: "This is a sample document for Platigo.",
}

response, err := client.Index(context.Background(), "index-name", doc)
if err != nil {
    log.Fatal("Failed to index document:", err)
}
```

**Creating an Index**
```go
indexName := "new-index"

mapping := `
{
    "mappings": {
        "properties": {
            "title": {
                "type": "text"
            },
            "content": {
                "type": "text"
            }
        }
    }
}
`

response, err := client.CreateIndices(context.Background(), indexName, strings.NewReader(mapping))
if err != nil {
    log.Fatal("Failed to create index:", err)
}
```

**Performing a Search**
```go
searchQuery := `
{
    "query": {
        "match": {
            "content": "sample"
        }
    }
}
`

response, err := client.Search(context.Background(), []string{"index-name"}, strings.NewReader(searchQuery))
if err != nil {
    log.Fatal("Failed to perform search:", err)
}
```

For more details on available utility functions and their usage, please refer to the [Platigo GitHub repository](https://github.com/bagastri07/platigo).

## Contribution

Contributions to Platigo are welcome! If you encounter any issues, have suggestions for improvements, or would like to contribute new features or enhancements, please feel free to open an issue

