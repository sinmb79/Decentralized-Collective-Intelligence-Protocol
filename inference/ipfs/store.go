package ipfs

import (
	"errors"
	"io"
	"os"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
)

const DefaultAPIURL = "http://127.0.0.1:5001"

type shellClient interface {
	Add(io.Reader, ...shell.AddOpts) (string, error)
	Cat(string) (io.ReadCloser, error)
}

// Client stores and retrieves inference payloads through the IPFS HTTP API.
type Client struct {
	apiURL string
	shell  shellClient
}

var newShell = func(apiURL string) shellClient {
	return shell.NewShell(apiURL)
}

// NewClient creates an IPFS client using the configured or default API URL.
func NewClient(apiURL string) *Client {
	url := strings.TrimSpace(apiURL)
	if url == "" {
		url = strings.TrimSpace(os.Getenv("IPFS_API_URL"))
	}
	if url == "" {
		url = DefaultAPIURL
	}

	return &Client{
		apiURL: strings.TrimRight(url, "/"),
		shell:  newShell(strings.TrimRight(url, "/")),
	}
}

// Store saves content to IPFS and returns its CID.
func Store(content string) (string, error) {
	return NewClient("").Store(content)
}

// Retrieve reads content from IPFS by CID.
func Retrieve(cid string) (string, error) {
	return NewClient("").Retrieve(cid)
}

// Store saves content to IPFS and returns its CID.
func (c *Client) Store(content string) (string, error) {
	if c == nil || c.shell == nil {
		return "", errors.New("ipfs client is not ready")
	}

	return c.shell.Add(strings.NewReader(content))
}

// Retrieve reads content from IPFS by CID.
func (c *Client) Retrieve(cid string) (string, error) {
	if c == nil || c.shell == nil {
		return "", errors.New("ipfs client is not ready")
	}

	resolvedCID := strings.TrimSpace(cid)
	if resolvedCID == "" {
		return "", errors.New("cid is empty")
	}

	reader, err := c.shell.Cat(resolvedCID)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
