package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"
)

var (
	ErrorNoLoginAuthenticationForUsername = errors.New("no login authentication info found for user_name")
)

// LoginAuthentication represents a location based event for tracked assets
type LoginAuthentication struct {
	bun.BaseModel `bun:"table:login_authentication"`
	ID            int64  `bun:",pk,autoincrement"`
	UserName      string `bun:"user_name"`
	PasswordHash  string `bun:"resource_id"`
}

// Save saves the current activity to
// the database, returning error (if any)
func (a *LoginAuthentication) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().Model(a).Exec(ctx)

	return err
}

// Load returns the row in the activities table
// returning error (if any)
func (a *LoginAuthentication) Load(ctx context.Context, db *bun.DB) error {
	return db.NewSelect().Model(a).WherePK().Scan(ctx)
}

// Update updates a row in the activities table
// returning error (if any)
func (a *LoginAuthentication) Update(ctx context.Context, db *bun.DB) error {
	_, err := db.NewUpdate().Model(a).WherePK().Exec(ctx)

	return err
}

// GetLoginAuthenticationsByUserName returns the row in the
// login_authentications table that has the specified userName
// or error if any of the below are true
// an error was encountered searching the database
// no login authentication information exists for that userName
func GetLoginAuthenticationsByUserName(ctx context.Context, db *bun.DB, userName string) (LoginAuthentication, error) {
	var loginAuthentications []LoginAuthentication
	err := db.NewSelect().Model(&loginAuthentications).Where("user_name = ?", userName).Scan(ctx)

	if err != nil {
		return LoginAuthentication{}, err
	}

	if len(loginAuthentications) != 1 {
		return LoginAuthentication{}, ErrorNoLoginAuthenticationForUsername
	}

	return loginAuthentications[0], nil
}

// ListLoginAuthenticationsWithPagination returns a page of login authentication rows up to limit size
// starting from the given cursor, returning 0 as the cursor if there are no more rows
func ListLoginAuthenticationsWithPagination(ctx context.Context, db *bun.DB, cursor int64, limit int) ([]LoginAuthentication, int64, error) {
	var loginAuthentications []LoginAuthentication
	var nextCursor int64
	// select one more than the limit to figure out if we have a next page
	err := db.NewSelect().Model(&loginAuthentications).Where("id > ?", cursor).OrderExpr("id ASC").Limit(limit + 1).Scan(ctx)

	if err != nil {
		if err != sql.ErrNoRows {
			return loginAuthentications, 0, err
		}
		// no rows, return no activities and nextCursor that indicates no more
		// pages to fetch
		return loginAuthentications, nextCursor, nil
	}

	if len(loginAuthentications) == 0 {
		// no rows, return no activities and nextCursor that indicates no more
		// pages to fetch
		return loginAuthentications, nextCursor, nil
	}

	// check if we are at the last page by seeing if there were a less or equal number
	//  of results then the max the user wants to fetch
	if len(loginAuthentications)-1 == limit {
		// look up the id of the last one
		// (minus the additional extra we fetched to see if there is a next page)
		// to use as the cursor for where to start fetching the next page of activities
		nextCursor = loginAuthentications[limit-2].ID
	}

	if len(loginAuthentications)-1 < limit {
		nextCursor = 0
	}

	// otherwise leave nextCursor as 0 to signal no more pages
	return loginAuthentications, nextCursor, err
}
