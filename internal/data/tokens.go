package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"time"

	"github.com/PedroDrago/greenlight/internal/validator"
)

type TokenModel struct {
	DB *sql.DB
}

type Token struct {
	Hash      []byte
	PlainText string
	UserID    int64
	Expiry    time.Time
	Scope     string
}

const (
	ScopeActivation = "activation"
)

func GenerateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]
	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

func (t *Token) Validate(v *validator.Validator) {
	ValidateTokenPlaintext(v, t.PlainText)
}

func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := GenerateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = m.Insert(token)
	return token, err
}

func (m *TokenModel) Insert(token *Token) error {
	query := `
    INSERT INTO tokens (hash, user_id, expiry, scope)
    VALUES  ($1, $2, $3, $4)`
	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
    DELETE FROM tokens
    WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}

func (m TokenModel) GetAllForUser(usr *User) ([]Token, error) {
	query := `
    SELECT hash, user_id, expiry, scope
    FROM tokens
    WHERE user_id = $1`

	var tokens []Token
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, usr.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	defer rows.Close()
	for rows.Next() {
		var token Token
		err := rows.Scan(&token.Hash, &token.UserID, &token.Expiry, &token.Scope)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}
