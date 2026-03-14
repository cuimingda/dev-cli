package cmd

import "testing"

func TestLoadGitHubAuthBaseConfigUsesDefaultAPIBaseURL(t *testing.T) {
	initializer := newGitHubLoginTestInitializer(t)

	config, err := loadGitHubAuthBaseConfig(initializer)
	if err != nil {
		t.Fatalf("loadGitHubAuthBaseConfig() returned error: %v", err)
	}

	if config.APIBaseURL != "https://api.github.com" {
		t.Fatalf("APIBaseURL = %q, want %q", config.APIBaseURL, "https://api.github.com")
	}
	if got := config.AuthBaseURL.String(); got != "https://github.com" {
		t.Fatalf("AuthBaseURL = %q, want %q", got, "https://github.com")
	}
	if config.Account != "github.com" {
		t.Fatalf("Account = %q, want %q", config.Account, "github.com")
	}
}

func TestLoadGitHubLoginConfigUsesDefaultAPIBaseURL(t *testing.T) {
	initializer := newGitHubLoginTestInitializer(t)
	if err := initializer.SetValue("github.client_id", "client-123"); err != nil {
		t.Fatalf("SetValue() returned error: %v", err)
	}

	config, err := loadGitHubLoginConfig(initializer)
	if err != nil {
		t.Fatalf("loadGitHubLoginConfig() returned error: %v", err)
	}

	if config.ClientID != "client-123" {
		t.Fatalf("ClientID = %q, want %q", config.ClientID, "client-123")
	}
	if config.APIBaseURL != "https://api.github.com" {
		t.Fatalf("APIBaseURL = %q, want %q", config.APIBaseURL, "https://api.github.com")
	}
	if got := config.AuthBaseURL.String(); got != "https://github.com" {
		t.Fatalf("AuthBaseURL = %q, want %q", got, "https://github.com")
	}
	if config.Account != "github.com" {
		t.Fatalf("Account = %q, want %q", config.Account, "github.com")
	}
}
