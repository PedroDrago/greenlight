package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/PedroDrago/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var ErrDuplicateEmail = errors.New("duplicate email")

type UserModel struct {
	DB *sql.DB
}
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int32     `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must not be empty")
	passwdLen := len(password)
	v.Check(passwdLen >= 8, "password", "must be at least 8 bytes long")
	v.Check(passwdLen <= 72, "password", "must not be more than 72 bytes long")
}

func (usr *User) Validate(v *validator.Validator) {
	v.Check(usr.Name != "", "name", "must be provided")
	v.Check(len(usr.Name) <= 500, "name", "must not be more than 500 bytes long")
	ValidateEmail(v, usr.Email)
	if usr.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *usr.Password.plaintext)
	}
	if usr.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func (m *UserModel) Insert(usr *User) error {
	query := `
    INSERT INTO users (name, email, password_hash, activated)
    VALUES ($1, $2, $3, $4)
    RETURNING id, created_at, version
    `
	args := []any{usr.Name, usr.Email, usr.Password.hash, usr.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&usr.ID, &usr.CreatedAt, &usr.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m *UserModel) Update(usr *User) error {
	query := `
    UPDATE users
    SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
    WHERE id = $5 AND version = $6
    RETURNING version
    `

	args := []any{
		usr.Name,
		usr.Email,
		usr.Password.hash,
		usr.Activated,
		usr.ID,
		usr.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&usr.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

func (m *UserModel) GetByEmail(email string) (*User, error) {
	query := `
    SELECT id, created_at, name, email. password_hash, activated, version
    FROM users
    WHERE email = $1
    `

	var usr User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&usr.ID,
		&usr.CreatedAt,
		&usr.Name,
		&usr.Email,
		&usr.Password.hash,
		&usr.Activated,
		&usr.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &usr, nil
}
