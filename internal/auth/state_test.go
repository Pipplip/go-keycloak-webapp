package auth

import (
	"testing"
	"time"
)

func TestStateStoreTakeRequiresMatchingBrowserAndIsOneTime(t *testing.T) {
	store := NewStateStore()

	transaction, err := store.Create("browser-binding")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, ok := store.Take(transaction.State, "different-browser"); ok {
		t.Fatal("Take() accepted a different browser binding")
	}

	got, ok := store.Take(transaction.State, "browser-binding")
	if !ok {
		t.Fatal("Take() rejected the matching browser binding")
	}
	if got != transaction {
		t.Fatalf("Take() transaction = %#v, want %#v", got, transaction)
	}

	if _, ok := store.Take(transaction.State, "browser-binding"); ok {
		t.Fatal("Take() accepted a transaction more than once")
	}
}

func TestStateStoreRejectsExpiredTransaction(t *testing.T) {
	store := NewStateStore()

	transaction, err := store.Create("browser-binding")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	store.mu.Lock()
	stored := store.values[transaction.State]
	stored.createdAt = time.Now().Add(-loginTransactionLifetime)
	store.values[transaction.State] = stored
	store.mu.Unlock()

	if _, ok := store.Take(transaction.State, "browser-binding"); ok {
		t.Fatal("Take() accepted an expired transaction")
	}
}

func TestStateStoreRejectsMissingBrowserBinding(t *testing.T) {
	store := NewStateStore()
	if _, err := store.Create(""); err == nil {
		t.Fatal("Create() accepted an empty browser binding")
	}
}
