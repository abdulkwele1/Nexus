package database

import (
	"context"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

var (
	ErrorNoLoginCookie = errors.New("no login cookie info found for user")
)

// LoginCookie represents a location based event for tracked assets
type LoginCookie struct {
	bun.BaseModel `bun:"table:login_cookies"`
	Cookie        string    `bun:",pk,cookie"`
	UserName      string    `bun:"user_name"`
	Expiration    time.Time `bun:"expires_at"`
}

// Save saves the current cookie to
// the database, returning error (if any)
func (lc *LoginCookie) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().Model(lc).Exec(ctx)

	return err
}

// Upsert inserts or updates the cookie for the user
// in the database, returning error (if any)
func (lc *LoginCookie) Upsert(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().On("CONFLICT (user_name) DO UPDATE").Model(lc).Exec(ctx)

	return err
}

// Load returns the row in the login_cookies table
// returning error (if any)
func (lc *LoginCookie) Load(ctx context.Context, db *bun.DB) error {
	return db.NewSelect().Model(lc).WherePK().Scan(ctx)
}

// Update updates a row in the login_cookies table
// returning error (if any)
func (lc *LoginCookie) Update(ctx context.Context, db *bun.DB) error {
	_, err := db.NewUpdate().Model(lc).WherePK().Exec(ctx)

	return err
}

// Delete deletes the cookie for the username
// returning error (if any)
func (lc *LoginCookie) Delete(ctx context.Context, db *bun.DB) error {
	_, err := db.NewDelete().Model(lc).WherePK().Exec(ctx)

	return err
}

// GetLoginCookieByUserName returns the row in the
// login_cookies table that has the specified cookie
// or error if any of the below are true
// an error was encountered searching the database
// no login cookie exists
func GetLoginCookie(ctx context.Context, db *bun.DB, cookie string) (LoginCookie, error) {
	var loginCookies []LoginCookie
	err := db.NewSelect().Model(&loginCookies).Where("cookie = ?", cookie).Scan(ctx)

	if err != nil {
		return LoginCookie{}, err
	}

	if len(loginCookies) != 1 {
		return LoginCookie{}, ErrorNoLoginCookie
	}

	return loginCookies[0], nil
}

func DeleteExpiredCookies(ctx context.Context, now time.Time, db *bun.DB) error {
	_, err := db.NewDelete().Model(&LoginCookie{}).Where("expires_at <= ?", now).Exec(ctx)

	return err
}
