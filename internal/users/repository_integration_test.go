package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"go-keycloak-webapp/internal/database"
	"go-keycloak-webapp/migrations"
)

func TestSQLRepositoryUpsertOIDCUser(t *testing.T) {
	databaseURL := os.Getenv("APP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("APP_TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	randomID := make([]byte, 16)
	if _, err := rand.Read(randomID); err != nil {
		t.Fatalf("generate test identity: %v", err)
	}
	suffix := hex.EncodeToString(randomID)
	identity := OIDCIdentity{
		Issuer:   "https://integration-test.invalid/" + suffix,
		Subject:  suffix,
		Username: "first-login",
		Email:    "first@example.invalid",
	}

	t.Cleanup(func() {
		if _, err := db.ExecContext(
			context.Background(),
			"DELETE FROM app_users WHERE oidc_issuer = $1 AND oidc_subject = $2",
			identity.Issuer,
			identity.Subject,
		); err != nil {
			t.Errorf("remove test user: %v", err)
		}
	})

	repository := NewSQLRepository(db)
	first, err := repository.UpsertOIDCUser(ctx, identity)
	if err != nil {
		t.Fatalf("create first-login user: %v", err)
	}

	identity.Username = "updated-profile"
	identity.Email = "updated@example.invalid"
	second, err := repository.UpsertOIDCUser(ctx, identity)
	if err != nil {
		t.Fatalf("update returning user: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("upsert created a second user: first id %d, second id %d", first.ID, second.ID)
	}
	if second.Username != identity.Username || second.Email != identity.Email {
		t.Fatalf("upsert returned stale profile data: %#v", second)
	}
}
