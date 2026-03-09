package ipfs

import (
	"errors"
	"io"
	"strings"
	"testing"

	shell "github.com/ipfs/go-ipfs-api"
)

type fakeShell struct {
	cid     string
	data    map[string]string
	addErr  error
	catErr  error
	lastAdd string
}

func (f *fakeShell) Add(reader io.Reader, _ ...shell.AddOpts) (string, error) {
	if f.addErr != nil {
		return "", f.addErr
	}

	payload, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	f.lastAdd = string(payload)
	if f.data == nil {
		f.data = make(map[string]string)
	}
	f.data[f.cid] = f.lastAdd
	return f.cid, nil
}

func (f *fakeShell) Cat(cid string) (io.ReadCloser, error) {
	if f.catErr != nil {
		return nil, f.catErr
	}

	return io.NopCloser(strings.NewReader(f.data[cid])), nil
}

func TestClientStoreAndRetrieve(t *testing.T) {
	mock := &fakeShell{
		cid:  "bafy-test-cid",
		data: make(map[string]string),
	}
	client := &Client{
		apiURL: DefaultAPIURL,
		shell:  mock,
	}

	cid, err := client.Store("collective intelligence")
	if err != nil {
		t.Fatalf("Store returned error: %v", err)
	}
	if cid != "bafy-test-cid" {
		t.Fatalf("unexpected cid: %q", cid)
	}
	if mock.lastAdd != "collective intelligence" {
		t.Fatalf("unexpected stored payload: %q", mock.lastAdd)
	}

	content, err := client.Retrieve(cid)
	if err != nil {
		t.Fatalf("Retrieve returned error: %v", err)
	}
	if content != "collective intelligence" {
		t.Fatalf("unexpected retrieved content: %q", content)
	}
}

func TestStoreUsesEnvironmentConfiguredURL(t *testing.T) {
	t.Setenv("IPFS_API_URL", "http://127.0.0.1:5009/")

	mock := &fakeShell{
		cid:  "bafy-env-cid",
		data: make(map[string]string),
	}

	var gotURL string
	originalFactory := newShell
	newShell = func(apiURL string) shellClient {
		gotURL = apiURL
		return mock
	}
	t.Cleanup(func() {
		newShell = originalFactory
	})

	cid, err := Store("from env")
	if err != nil {
		t.Fatalf("Store returned error: %v", err)
	}
	if cid != "bafy-env-cid" {
		t.Fatalf("unexpected cid: %q", cid)
	}
	if gotURL != "http://127.0.0.1:5009" {
		t.Fatalf("unexpected API URL: %q", gotURL)
	}
}

func TestRetrieveRejectsEmptyCID(t *testing.T) {
	client := &Client{
		apiURL: DefaultAPIURL,
		shell:  &fakeShell{},
	}

	_, err := client.Retrieve("   ")
	if err == nil {
		t.Fatal("expected empty CID to return an error")
	}
}

func TestClientPropagatesShellErrors(t *testing.T) {
	client := &Client{
		apiURL: DefaultAPIURL,
		shell: &fakeShell{
			cid:    "bafy-fail",
			addErr: errors.New("ipfs offline"),
		},
	}

	_, err := client.Store("payload")
	if err == nil || err.Error() != "ipfs offline" {
		t.Fatalf("expected shell error, got %v", err)
	}
}
