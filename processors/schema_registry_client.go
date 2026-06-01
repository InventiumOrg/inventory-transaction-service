package processors

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// SchemaRegistryClient is a minimal Confluent Schema Registry HTTP client
// supporting the subset of operations needed by the producer.
type SchemaRegistryClient struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
}

// SchemaResponse is the shape returned by `/subjects/.../versions/...` and
// `/schemas/ids/{id}` endpoints.
type SchemaResponse struct {
	Subject string `json:"subject,omitempty"`
	Version int    `json:"version,omitempty"`
	ID      int    `json:"id"`
	Schema  string `json:"schema"`
}

func NewSchemaRegistryClient(baseURL, username, password string) *SchemaRegistryClient {
	return &SchemaRegistryClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		username:   username,
		password:   password,
	}
}

func (c *SchemaRegistryClient) do(req *http.Request) ([]byte, error) {
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	req.Header.Set("Accept", "application/vnd.schemaregistry.v1+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("schema registry request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Error("Failed to close schema registry response body", slog.Any("error", closeErr))
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read schema registry body: %w", err)
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("schema registry returned %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// GetLatestSchema fetches the latest registered schema for the supplied subject.
func (c *SchemaRegistryClient) GetLatestSchema(subject string) (*SchemaResponse, error) {
	endpoint := fmt.Sprintf("%s/subjects/%s/versions/latest", c.baseURL, url.PathEscape(subject))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	body, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var out SchemaResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode schema response: %w", err)
	}
	return &out, nil
}

// GetSchemaByID fetches a schema by its global ID.
func (c *SchemaRegistryClient) GetSchemaByID(id int) (string, error) {
	endpoint := fmt.Sprintf("%s/schemas/ids/%d", c.baseURL, id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	body, err := c.do(req)
	if err != nil {
		return "", err
	}
	var out SchemaResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("decode schema response: %w", err)
	}
	return out.Schema, nil
}

type registerRequest struct {
	Schema     string `json:"schema"`
	SchemaType string `json:"schemaType,omitempty"`
}

// RegisterSchema registers (or returns an existing) schema for the supplied
// subject and returns its global ID.
func (c *SchemaRegistryClient) RegisterSchema(subject, schema string) (int, error) {
	endpoint := fmt.Sprintf("%s/subjects/%s/versions", c.baseURL, url.PathEscape(subject))
	payload, err := json.Marshal(registerRequest{Schema: schema, SchemaType: "AVRO"})
	if err != nil {
		return 0, fmt.Errorf("marshal register request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Body = io.NopCloser(bytesReader(payload))
	req.ContentLength = int64(len(payload))
	req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")

	body, err := c.do(req)
	if err != nil {
		return 0, err
	}
	var out struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return 0, fmt.Errorf("decode register response: %w", err)
	}
	return out.ID, nil
}

// bytesReader returns an io.Reader for the supplied byte slice without
// pulling in `bytes` purely for this purpose.
func bytesReader(b []byte) io.Reader {
	return &sliceReader{data: b}
}

type sliceReader struct {
	data []byte
	pos  int
}

func (r *sliceReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
