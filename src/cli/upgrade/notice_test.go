package upgrade

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	rhttp "github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
	"github.com/stretchr/testify/assert"
)

// testRoundTripper redirects all outbound requests to a local test server.
type testRoundTripper struct {
	target *url.URL
}

func (rt *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.URL = &url.URL{
		Scheme: rt.target.Scheme,
		Host:   rt.target.Host,
		Path:   req.URL.Path,
	}
	return http.DefaultTransport.RoundTrip(cloned)
}

func TestCanUpgrade(t *testing.T) {
	const fakeLatest = "99.99.99"

	// Serve a fake version file locally so tests never hit GitHub.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "v%s", fakeLatest)
	}))
	defer server.Close()

	targetURL, _ := url.Parse(server.URL)

	savedHTTPClient := rhttp.HTTPClient
	rhttp.HTTPClient = &http.Client{Transport: &testRoundTripper{target: targetURL}}
	defer func() { rhttp.HTTPClient = savedHTTPClient }()

	savedIsConnected := rhttp.IsConnected
	rhttp.IsConnected = func() bool { return true }
	defer func() { rhttp.IsConnected = savedIsConnected }()

	ugc := &Config{}
	latest, err := ugc.FetchLatest()
	if err != nil {
		t.Fatalf("failed to fetch latest version: %v", err)
	}

	cases := []struct {
		Case           string
		CurrentVersion string
		Installer      string
		Expected       bool
		Cache          bool
	}{
		{Case: "Up to date", CurrentVersion: latest},
		{Case: "Outdated Linux", Expected: true, CurrentVersion: "3.0.0"},
		{Case: "Outdated Darwin", Expected: true, CurrentVersion: "3.0.0"},
		{Case: "Cached", Cache: true, CurrentVersion: latest},
		{Case: "Windows Store", Installer: "ws"},
	}

	for _, tc := range cases {
		build.Version = tc.CurrentVersion

		if len(tc.Installer) > 0 {
			os.Setenv("POSH_INSTALLER", tc.Installer)
		}

		_, canUpgrade := ugc.Notice()
		assert.Equal(t, tc.Expected, canUpgrade, tc.Case)

		os.Setenv("POSH_INSTALLER", "")
	}
}
