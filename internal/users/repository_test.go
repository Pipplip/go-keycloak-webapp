package users

import (
	"context"
	"testing"
)

func TestUpsertOIDCUserRequiresIssuerAndSubject(t *testing.T) {
	repository := NewSQLRepository(nil)

	for _, identity := range []OIDCIdentity{
		{Subject: "subject"},
		{Issuer: "https://issuer.example/realm"},
	} {
		if _, err := repository.UpsertOIDCUser(context.Background(), identity); err == nil {
			t.Fatalf("UpsertOIDCUser(%#v) accepted an incomplete identity", identity)
		}
	}
}
