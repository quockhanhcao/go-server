package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCheckPasswordHash(t *testing.T) {
	hash, err := HashPassword("testpassword")
	require.NoError(t, err)
	require.NotNil(t, hash)
	err = CheckPasswordHash(hash, "testpassword")
	require.NoError(t, err)
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "secret", time.Hour)

	// valid token
	gotUserID, err := ValidateJWT(validToken, "secret")
	require.NoError(t, err)
	require.Equal(t, userID, gotUserID)

	// invalid token string
	gotUserID, err = ValidateJWT("invalid.token.string", "secret")
	require.Error(t, err)
	require.Equal(t, uuid.Nil, gotUserID)
	require.NotEqual(t, userID, gotUserID)

	// wrong secret
	gotUserID, err = ValidateJWT(validToken, "wrong_secret")
	require.Error(t, err)
	require.Equal(t, uuid.Nil, gotUserID)
	require.NotEqual(t, userID, gotUserID)
}

func TestGetBearerToken(t *testing.T) {
	// valid token
	headers := http.Header{
		"Authorization": []string{"Bearer abcdef123456"},
	}
	token, err := GetBearerToken(headers)
	require.NoError(t, err)
	require.Equal(t, "abcdef123456", token)

	// no authorization token
	headers = http.Header{
		"Content-type": []string{"Application/json"},
	}
	token, err = GetBearerToken(headers)
	require.Error(t, err, "authorization header is missing")
	require.Equal(t, token, "")

	// invalid authorization header
	headers = http.Header{
		"Authorization": []string{"Bearer abcdef123456 absceariS"},
	}
	token, err = GetBearerToken(headers)
	require.Error(t, err, "invalid Authorization header format: Bearer abcdef123456 absceariS")
	require.Equal(t, token, "")

	// authorization header does not contain keyword bearer
	headers = http.Header{
		"Authorization": []string{"Dearer abcdef123456 absceariS"},
	}
	token, err = GetBearerToken(headers)
	require.Error(t, err, "invalid Authorization header format: Dearer abcdef123456 absceariS")
	require.Equal(t, token, "")
}
