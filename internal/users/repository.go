package users

import (
	"context"
	"database/sql"
	"fmt"
)

type OIDCIdentity struct {
	Issuer   string
	Subject  string
	Username string
	Email    string
}

type User struct {
	ID       int64
	Issuer   string
	Subject  string
	Username string
	Email    string
}

type Repository interface {
	UpsertOIDCUser(context.Context, OIDCIdentity) (User, error)
}

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) UpsertOIDCUser(
	ctx context.Context,
	identity OIDCIdentity,
) (User, error) {
	if identity.Issuer == "" || identity.Subject == "" {
		return User{}, fmt.Errorf("OIDC issuer and subject are required")
	}

	var user User
	err := r.db.QueryRowContext(
		ctx,
		`
			INSERT INTO app_users (oidc_issuer, oidc_subject, username, email)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (oidc_issuer, oidc_subject)
			DO UPDATE SET
				username = EXCLUDED.username,
				email = EXCLUDED.email,
				updated_at = CURRENT_TIMESTAMP
			RETURNING id, oidc_issuer, oidc_subject, username, email
		`,
		identity.Issuer,
		identity.Subject,
		identity.Username,
		identity.Email,
	).Scan(&user.ID, &user.Issuer, &user.Subject, &user.Username, &user.Email)
	if err != nil {
		return User{}, fmt.Errorf("upsert OIDC user: %w", err)
	}

	return user, nil
}
