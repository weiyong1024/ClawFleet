package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAssetsMarksLegacySlackAssetsUnvalidated(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	dataDir := filepath.Join(tempHome, ".clawfleet")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	assetsJSON := `{
  "channels": [
    {
      "id": "legacy-slack",
      "name": "Slack Bot",
      "channel": "slack",
      "token": "xoxb-legacy",
      "validated": true
    },
    {
      "id": "telegram",
      "name": "Telegram Bot",
      "channel": "telegram",
      "token": "123456:ABC",
      "validated": true
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(dataDir, "assets.json"), []byte(assetsJSON), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	store, err := LoadAssets()
	if err != nil {
		t.Fatalf("LoadAssets failed: %v", err)
	}

	slack := store.GetChannel("legacy-slack")
	if slack == nil {
		t.Fatal("expected legacy Slack asset to exist")
	}
	if slack.Validated {
		t.Fatal("expected legacy Slack asset without app_token to be marked unvalidated")
	}

	telegram := store.GetChannel("telegram")
	if telegram == nil {
		t.Fatal("expected Telegram asset to exist")
	}
	if !telegram.Validated {
		t.Fatal("expected non-Slack assets to keep their validated state")
	}
}
