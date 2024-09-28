package password

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnitTestCheckPasswordHashReturnsTrueWhenCorrectHashForPassword(t *testing.T) {
	// setup test data
	testPassword := "password123"
	testHash := "$2a$10$HqQx4jxUzfQm1fZYUZRLbOBaMNWHmhSmweH03rl0EykgE4BNfDciO"

	// execute test
	match := CheckPasswordHash(testPassword, testHash)

	// assert results
	assert.True(t, match, fmt.Sprintf("expected hash %s to match for password %s", testHash, testPassword))
}

func TestUnitTestCheckPasswordHashReturnsFalseWhenInCorrectHashForPassword(t *testing.T) {
	// setup test data
	testPassword := "password123"
	testHash := "WRONGHash"

	// execute test
	match := CheckPasswordHash(testPassword, testHash)

	// assert results
	assert.False(t, match, fmt.Sprintf("expected hash %s not to match for password %s", testHash, testPassword))
}

func TestUnitTestCheckPasswordHashReturnsFalseWhenIncorrectHashForPassword(t *testing.T) {
	// setup test data
	testPassword := "differentPasswordThenHash"
	testHash := "$2a$10$HqQx4jxUzfQm1fZYUZRLbOBaMNWHmhSmweH03rl0EykgE4BNfDciO"

	// execute test
	match := CheckPasswordHash(testPassword, testHash)

	// assert results
	assert.False(t, match, fmt.Sprintf("expected hash %s not to match for password %s", testHash, testPassword))
}

func TestUnitTestHashPasswordResultIsValidHashForPassword(t *testing.T) {
	// setup test data
	randomPassword := uuid.NewString()
	hashForRandomPassword, err := HashPassword(randomPassword)

	assert.NoError(t, err)

	// execute test
	match := CheckPasswordHash(randomPassword, hashForRandomPassword)

	// assert test expectations
	assert.True(t, match, fmt.Sprintf("expected hash %s to match for password %s", hashForRandomPassword, randomPassword))

}
