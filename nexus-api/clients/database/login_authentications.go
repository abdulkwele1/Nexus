package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/uptrace/bun"
)

var (
	ErrorNoLoginAuthenticationForUsername = errors.New("no login authentication info found for user_name")
)

type LoginAuthentication struct {
	bun.BaseModel `bun:"table:login_authentications"`
	ID            int64  `bun:",pk,autoincrement"`
	UserName      string `bun:"user_name"`
	PasswordHash  string `bun:"password_hash"`
	Role          string `bun:"role"`
}

func (a *LoginAuthentication) Save(ctx context.Context, db *bun.DB) error {
	_, err := db.NewInsert().Model(a).Exec(ctx)

	return err
}

func (a *LoginAuthentication) Load(ctx context.Context, db *bun.DB) error {
	return db.NewSelect().Model(a).WherePK().Scan(ctx)
}

func (a *LoginAuthentication) Update(ctx context.Context, db *bun.DB) error {
	_, err := db.NewUpdate().Model(a).WherePK().Exec(ctx)

	return err
}

// GetLoginAuthenticationByUserName returns the row in the
// login_authentications table that has the specified userName
// or error if any of the below are true
// an error was encountered searching the database
// no login authentication information exists for that userName
func GetLoginAuthenticationByUserName(ctx context.Context, db *bun.DB, userName string) (LoginAuthentication, error) {
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

// GetAllUsers returns all users with their roles and basic info
func GetAllUsers(ctx context.Context, db *bun.DB) ([]LoginAuthentication, error) {
	var users []LoginAuthentication
	err := db.NewSelect().Model(&users).OrderExpr("user_name ASC").Scan(ctx)
	return users, err
}

// GetUserRole returns the role of a specific user
func GetUserRole(ctx context.Context, db *bun.DB, username string) (string, error) {
	var user LoginAuthentication
	err := db.NewSelect().
		Model(&user).
		Column("role").
		Where("user_name = ?", username).
		Scan(ctx)
	if err != nil {
		return "", err
	}
	return user.Role, nil
}

// UpdateUserRole updates the role of a specific user
func UpdateUserRole(ctx context.Context, db *bun.DB, username, role string) error {
	// Validate role
	if role != "user" && role != "admin" && role != "root_admin" {
		return fmt.Errorf("invalid role: %s", role)
	}

	_, err := db.NewUpdate().
		Model((*LoginAuthentication)(nil)).
		Set("role = ?", role).
		Where("user_name = ?", username).
		Exec(ctx)

	return err
}
