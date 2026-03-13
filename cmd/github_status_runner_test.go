package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGitHubAuthStatusRunnerRunReportsNotLoggedInWhenTokenIsMissing(t *testing.T) {
	initializer := newGitHubLoginTestInitializer(t)
	if err := initializer.SetValue("github.client_id", "client-123"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	runner := &GitHubAuthStatusRunner{
		initializer: initializer,
		httpClient:  http.DefaultClient,
		now: func() time.Time {
			return time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
		},
		tokenStore: &stubGitHubTokenStore{
			loadErr: ErrGitHubTokenNotFound,
		},
		expiringSoonThreshold: githubAccessTokenExpiringSoonThreshold,
	}

	var output bytes.Buffer
	if err := runner.Run(context.Background(), &output); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	reportOutput := output.String()
	if !strings.Contains(reportOutput, "Status: Not logged in: no access token is stored in the macOS keychain.") {
		t.Fatalf("output = %q, want not logged in status", reportOutput)
	}
	if !strings.Contains(reportOutput, "Recommended next step: run `dev github login`") {
		t.Fatalf("output = %q, want login recommendation", reportOutput)
	}
	if !strings.Contains(reportOutput, "- GET /user: skipped because no access token is stored locally") {
		t.Fatalf("output = %q, want skipped remote probe", reportOutput)
	}
}

func TestGitHubAuthStatusRunnerEvaluateReturnsValidStateWhenRemoteProbeSucceeds(t *testing.T) {
	initializer := newGitHubLoginTestInitializer(t)
	if err := initializer.SetValue("github.client_id", "client-123"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/user" {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Fatalf("Authorization header = %q, want %q", got, "Bearer access-token")
		}

		writeGitHubJSONResponse(t, w, `{"login":"octocat"}`)
	}))
	defer server.Close()

	if err := initializer.SetValue("github.api_base_url", server.URL+"/api/v3"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	now := time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
	runner := &GitHubAuthStatusRunner{
		initializer: initializer,
		httpClient:  server.Client(),
		now: func() time.Time {
			return now
		},
		tokenStore: &stubGitHubTokenStore{
			loadToken: GitHubStoredToken{
				AccessToken:           "access-token",
				RefreshToken:          "refresh-token",
				AccessTokenExpiresAt:  timePointer(now.Add(2 * time.Hour)),
				RefreshTokenExpiresAt: timePointer(now.Add(24 * time.Hour)),
			},
		},
		expiringSoonThreshold: githubAccessTokenExpiringSoonThreshold,
	}

	report, err := runner.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate() returned error: %v", err)
	}

	if report.State != GitHubAuthStateTokenValid {
		t.Fatalf("report.State = %q, want %q", report.State, GitHubAuthStateTokenValid)
	}
	if report.NextAction != GitHubAuthNextActionCallAPI {
		t.Fatalf("report.NextAction = %q, want %q", report.NextAction, GitHubAuthNextActionCallAPI)
	}
	if report.Username != "octocat" {
		t.Fatalf("report.Username = %q, want %q", report.Username, "octocat")
	}
	if report.RemoteProbeStatusCode != http.StatusOK {
		t.Fatalf("report.RemoteProbeStatusCode = %d, want %d", report.RemoteProbeStatusCode, http.StatusOK)
	}
}

func TestGitHubAuthStatusRunnerEvaluateReturnsRefreshableStateWhenAccessTokenExpired(t *testing.T) {
	initializer := newGitHubLoginTestInitializer(t)
	if err := initializer.SetValue("github.client_id", "client-123"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	now := time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
	runner := &GitHubAuthStatusRunner{
		initializer: initializer,
		httpClient:  http.DefaultClient,
		now: func() time.Time {
			return now
		},
		tokenStore: &stubGitHubTokenStore{
			loadToken: GitHubStoredToken{
				AccessToken:           "access-token",
				RefreshToken:          "refresh-token",
				AccessTokenExpiresAt:  timePointer(now.Add(-1 * time.Minute)),
				RefreshTokenExpiresAt: timePointer(now.Add(24 * time.Hour)),
			},
		},
		expiringSoonThreshold: githubAccessTokenExpiringSoonThreshold,
	}

	report, err := runner.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate() returned error: %v", err)
	}

	if report.State != GitHubAuthStateAccessTokenExpiredRefreshable {
		t.Fatalf("report.State = %q, want %q", report.State, GitHubAuthStateAccessTokenExpiredRefreshable)
	}
	if report.NextAction != GitHubAuthNextActionRefresh {
		t.Fatalf("report.NextAction = %q, want %q", report.NextAction, GitHubAuthNextActionRefresh)
	}
	if report.RemoteProbeState != GitHubRemoteProbeSkipped {
		t.Fatalf("report.RemoteProbeState = %q, want %q", report.RemoteProbeState, GitHubRemoteProbeSkipped)
	}
}

func TestGitHubAuthStatusRunnerEvaluateReturnsAuthorizationInvalidOn401(t *testing.T) {
	initializer := newGitHubLoginTestInitializer(t)
	if err := initializer.SetValue("github.client_id", "client-123"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		writeGitHubJSONResponse(t, w, `{"message":"Bad credentials"}`)
	}))
	defer server.Close()

	if err := initializer.SetValue("github.api_base_url", server.URL+"/api/v3"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	now := time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
	runner := &GitHubAuthStatusRunner{
		initializer: initializer,
		httpClient:  server.Client(),
		now: func() time.Time {
			return now
		},
		tokenStore: &stubGitHubTokenStore{
			loadToken: GitHubStoredToken{
				AccessToken:          "access-token",
				AccessTokenExpiresAt: timePointer(now.Add(2 * time.Hour)),
			},
		},
		expiringSoonThreshold: githubAccessTokenExpiringSoonThreshold,
	}

	report, err := runner.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate() returned error: %v", err)
	}

	if report.State != GitHubAuthStateAuthorizationInvalid {
		t.Fatalf("report.State = %q, want %q", report.State, GitHubAuthStateAuthorizationInvalid)
	}
	if report.NextAction != GitHubAuthNextActionLogin {
		t.Fatalf("report.NextAction = %q, want %q", report.NextAction, GitHubAuthNextActionLogin)
	}
	if report.RemoteProbeStatusCode != http.StatusUnauthorized {
		t.Fatalf("report.RemoteProbeStatusCode = %d, want %d", report.RemoteProbeStatusCode, http.StatusUnauthorized)
	}
}

func TestGitHubAuthStatusRunnerEvaluateReturnsReauthenticationRequiredWhenRefreshTokenExpired(t *testing.T) {
	initializer := newGitHubLoginTestInitializer(t)
	if err := initializer.SetValue("github.client_id", "client-123"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	now := time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
	runner := &GitHubAuthStatusRunner{
		initializer: initializer,
		httpClient:  http.DefaultClient,
		now: func() time.Time {
			return now
		},
		tokenStore: &stubGitHubTokenStore{
			loadToken: GitHubStoredToken{
				AccessToken:           "access-token",
				RefreshToken:          "refresh-token",
				AccessTokenExpiresAt:  timePointer(now.Add(-1 * time.Minute)),
				RefreshTokenExpiresAt: timePointer(now.Add(-1 * time.Minute)),
			},
		},
		expiringSoonThreshold: githubAccessTokenExpiringSoonThreshold,
	}

	report, err := runner.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("Evaluate() returned error: %v", err)
	}

	if report.State != GitHubAuthStateReauthenticationRequired {
		t.Fatalf("report.State = %q, want %q", report.State, GitHubAuthStateReauthenticationRequired)
	}
	if report.NextAction != GitHubAuthNextActionLogin {
		t.Fatalf("report.NextAction = %q, want %q", report.NextAction, GitHubAuthNextActionLogin)
	}
}

func TestGitHubAuthStatusRunnerRunRecommendsFixingConfigWhenClientIDIsMissing(t *testing.T) {
	initializer := newGitHubLoginTestInitializer(t)

	runner := &GitHubAuthStatusRunner{
		initializer: initializer,
		httpClient:  http.DefaultClient,
		now: func() time.Time {
			return time.Date(2026, time.March, 13, 12, 0, 0, 0, time.UTC)
		},
		tokenStore: &stubGitHubTokenStore{
			loadErr: ErrGitHubTokenNotFound,
		},
		expiringSoonThreshold: githubAccessTokenExpiringSoonThreshold,
	}

	var output bytes.Buffer
	if err := runner.Run(context.Background(), &output); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if !strings.Contains(output.String(), "Recommended next step: configure `github.client_id` before logging in or refreshing") {
		t.Fatalf("output = %q, want config recommendation", output.String())
	}
}

func timePointer(value time.Time) *time.Time {
	return &value
}
