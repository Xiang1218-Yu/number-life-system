package service

import (
	"encoding/json"
	"strings"
	"testing"

	"number-life-system/internal/domain"
)

func TestBug3ExportKeepsAccountRelations(t *testing.T) {
	accountID := uint(42)
	bundle := domain.ImportBundle{
		Accounts: []domain.Account{{ID: accountID, Platform: "Example", Username: "linked-user"}},
		Subscriptions: []domain.Subscription{{
			AccountID:   &accountID,
			ServiceName: "Example Plus",
		}},
		Footprints: []domain.DigitalFootprint{{
			AccountID: &accountID,
			Title:     "登录",
		}},
		DataLocations: []domain.DataLocation{{
			AccountID: &accountID,
			Platform:  "Example",
		}},
		Backups: []domain.BackupRecord{{
			AccountID: &accountID,
			Platform:  "Example",
		}},
	}

	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"account_id":42`) {
		t.Fatalf("export bundle is missing account relations: %s", data)
	}
}
