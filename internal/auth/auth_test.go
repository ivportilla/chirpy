package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	secret := "my_secret"
	t.Run("Create a JWT token correctly", func(t *testing.T) {
		_, err := MakeJWT(uuid.New(), secret, 0*time.Second)
		if err != nil {
			t.Errorf("Unexpected error creating a valid token: %v", err)
		}
	})

	t.Run("Parse a JWT token correctly", func(t *testing.T) {
		token, _ := MakeJWT(uuid.New(), secret, 100*time.Second)
		_, err := ValidateJWT(token, secret)
		if err != nil {
			t.Errorf("Unexpected error validating a valid token: %v", err)
		}
	})

	t.Run("Handle an invalid JWT token correctly", func(t *testing.T) {
		_, err := ValidateJWT("asdasd", secret)
		if err == nil {
			t.Errorf("It should generate an error when used with an invalid JWT token")
		}
	})

	t.Run("Handle an expired JWT token correctly", func(t *testing.T) {
		// Create expired token
		token, _ := MakeJWT(uuid.New(), secret, -1*time.Second)
		_, err := ValidateJWT(token, secret)
		if err == nil {
			t.Errorf("It should generate an error when used with an expired JWT token")
		}
	})

	t.Run("Reject a token signed with a different secret", func(t *testing.T) {
		// Create expired token
		token, _ := MakeJWT(uuid.New(), "another secret", 100*time.Second)
		_, err := ValidateJWT(token, secret)
		if err == nil {
			t.Errorf("It should generate an error due to invalid secret used")
		}
	})
}

func TestGetBearerToken(t *testing.T) {
	t.Run("Extract a token correctly", func(t *testing.T) {
		header := http.Header{}
		expectedToken := "expected_token"
		header.Add("Authorization", "Bearer "+expectedToken)
		r, err := GetBearerToken(header)
		if err != nil {
			t.Errorf("Non error expected for a valid authorization header")
		}

		if r != expectedToken {
			t.Errorf("Expected token %s, but got %s", expectedToken, r)
		}
	})

	t.Run("Get an error when authorization is empty", func(t *testing.T) {
		header := http.Header{}
		header.Add("Authorization", "")
		_, err := GetBearerToken(header)
		if err == nil {
			t.Errorf("Expected error with empty auth header")
		}
	})

	t.Run("Get an error when authorization is invalid", func(t *testing.T) {
		header := http.Header{}
		header.Add("Authorization", "invalid auth header")
		_, err := GetBearerToken(header)
		if err == nil {
			t.Errorf("Expected error with empty auth header")
		}
	})
}

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
