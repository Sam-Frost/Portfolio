package spotify

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Sam-Frost/portfolio/internal/apperr"
)

const singletonID = "singleton"

type PostgresRepository struct {
	db     *sql.DB
	cipher *Cipher
}

func NewPostgresRepository(db *sql.DB, cipher *Cipher) *PostgresRepository {
	return &PostgresRepository{db: db, cipher: cipher}
}

func (r *PostgresRepository) Get(ctx context.Context) (TokenSet, bool, error) {
	const q = `SELECT refresh_token_cipher, access_token, access_token_expires_at FROM spotify_tokens WHERE id = $1`

	var cipherText, accessToken string
	var expiresAt time.Time
	err := r.db.QueryRowContext(ctx, q, singletonID).Scan(&cipherText, &accessToken, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return TokenSet{}, false, nil
	}
	if err != nil {
		return TokenSet{}, false, apperr.Internal("failed to load spotify tokens")
	}

	refreshToken, err := r.cipher.Decrypt(cipherText)
	if err != nil {
		return TokenSet{}, false, apperr.Internal("failed to decrypt spotify refresh token")
	}

	return TokenSet{
		RefreshToken:         refreshToken,
		AccessToken:          accessToken,
		AccessTokenExpiresAt: expiresAt,
	}, true, nil
}

func (r *PostgresRepository) Save(ctx context.Context, tokens TokenSet) error {
	encrypted, err := r.cipher.Encrypt(tokens.RefreshToken)
	if err != nil {
		return apperr.Internal("failed to encrypt spotify refresh token")
	}

	const q = `
		INSERT INTO spotify_tokens (id, refresh_token_cipher, access_token, access_token_expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			refresh_token_cipher = EXCLUDED.refresh_token_cipher,
			access_token = EXCLUDED.access_token,
			access_token_expires_at = EXCLUDED.access_token_expires_at
	`
	if _, err := r.db.ExecContext(ctx, q, singletonID, encrypted, tokens.AccessToken, tokens.AccessTokenExpiresAt); err != nil {
		return apperr.Internal("failed to save spotify tokens")
	}
	return nil
}
