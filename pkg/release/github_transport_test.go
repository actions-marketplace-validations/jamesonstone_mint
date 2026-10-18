package release

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReleaseRejectsUnsafeAPIURLsBeforeSendingToken(t *testing.T) {
	for _, apiURL := range []string{"http://api.github.com", "ftp://localhost", "https://user:password@api.github.com", "/relative"} {
		t.Run(apiURL, func(t *testing.T) {
			_, err := PublishGitHubRelease(context.Background(), GitHubReleaseOptions{
				Owner: "acme", Repo: "app", Tag: "v1.0.0", Target: "HEAD", Token: "test-token", APIBaseURL: apiURL,
			})
			if err == nil || !strings.Contains(err.Error(), "HTTPS") {
				t.Fatalf("expected transport rejection, got %v", err)
			}
		})
	}
}

func TestReleaseRefusesRedirectsIncludingCustomClient(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(map[bool]string{false: "default", true: "custom"}[custom], func(t *testing.T) {
			received := false
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { received = true; w.WriteHeader(http.StatusUnauthorized) }))
			defer target.Close()
			source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
			}))
			defer source.Close()
			var client *http.Client
			if custom {
				client = &http.Client{}
			}
			_, err := PublishGitHubRelease(context.Background(), GitHubReleaseOptions{
				Owner: "acme", Repo: "app", Tag: "v1.0.0", Target: "HEAD", Token: "test-token", APIBaseURL: source.URL, HTTPClient: client,
			})
			if err == nil || !strings.Contains(err.Error(), "307") {
				t.Fatalf("expected redirect rejection, got %v", err)
			}
			if received {
				t.Fatal("redirect target received a credential-bearing request")
			}
			if custom && client.CheckRedirect != nil {
				t.Fatal("caller's client was mutated")
			}
		})
	}
}
