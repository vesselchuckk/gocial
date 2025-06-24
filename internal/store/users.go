package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	exp                  = time.Hour * 24 * 3
	ErrDuplicateEmail    = errors.New("a user with this email already exists")
	ErrDuplicateUsername = errors.New("a user with this username already exists")
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"password_hash" db:"password_hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	RoleID    int64     `json:"role_id" db:"role_id"`
	Role      Role      `json:"role" db:"level"`
	RoleName  string    `json:"name" db:"name"`
}

type UsersStore struct {
	db *sqlx.DB
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func NewUsersStore(db *sql.DB) *UsersStore {
	return &UsersStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

func (s *UsersStore) CreateUser(ctx context.Context, tx *sqlx.Tx, user *User) error {
	const query = `INSERT INTO users (username, email, password_hash, role_id) VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4)) RETURNING *;`

	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash the password: %w", err)
	}
	passhash := base64.StdEncoding.EncodeToString(bytes)

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	role := user.Role.Name
	if role == "" {
		role = "user"
	}

	if err := tx.GetContext(ctx, user, query, user.Username, user.Email, passhash, role); err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (s *UsersStore) GetByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	const query = `SELECT users.id,
			users.username,
			users.email,
			users.password_hash,
			users.created_at,
			users.is_active,
			users.role_id,
			roles.name as name
		FROM users JOIN roles ON (users.role_id = roles.id) WHERE users.id = $1 AND users.is_active=true;`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User
	if err := s.db.GetContext(ctx, &user, query, userID); err != nil {
		return nil, fmt.Errorf("failed to get user from DB: %w", err)
	}

	return &user, nil
}

func (s *UsersStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	const query = "SELECT * FROM users WHERE email = $1 AND is_active = true;"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var user User
	if err := s.db.GetContext(ctx, &user, query, email); err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (s *UsersStore) delete(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) error {
	const query = `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (s *UsersStore) Delete(ctx context.Context, userID uuid.UUID) error {
	return withTx(s.db, ctx, func(tx *sqlx.Tx) error {
		if err := s.delete(ctx, tx, userID); err != nil {
			return err
		}

		if err := s.deleteInvite(ctx, tx, userID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UsersStore) createUserInvitation(ctx context.Context, tx *sqlx.Tx, token string, exp time.Duration, userID uuid.UUID) error {
	const query = `INSERT INTO invitations (token, user_id, expiry) VALUES ($1, $2, $3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (s *UsersStore) CreateAndInvite(ctx context.Context, user *User, token string, invExp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sqlx.Tx) error {
		if err := s.CreateUser(ctx, tx, user); err != nil {
			return err
		}

		if err := s.createUserInvitation(ctx, tx, token, invExp, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UsersStore) getUserFromInvitation(ctx context.Context, tx *sqlx.Tx, token string) (*User, error) {
	const query = `SELECT u.id, u.username, u.email, u.created_at, u.is_active
				   FROM users u
				   JOIN invitations ui ON u.id = ui.user_id
				   WHERE ui.token = $1 AND ui.expiry > $2`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	if err := tx.GetContext(ctx, user, query, hashToken, time.Now()); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UsersStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sqlx.Tx) error {
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}

		if err := s.deleteInvite(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UsersStore) update(ctx context.Context, tx *sqlx.Tx, user *User) error {

	const query = `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	if err := s.deleteInvite(ctx, tx, user.ID); err != nil {
		return err
	}

	return nil
}

func (s *UsersStore) deleteInvite(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) error {
	const query = `DELETE FROM invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}
