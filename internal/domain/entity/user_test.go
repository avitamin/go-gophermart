package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserCredentials_Validate(t *testing.T) {
	tests := []struct {
		name    string
		creds   UserCredentials
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid credentials",
			creds: UserCredentials{
				Login:    "testuser",
				Password: "12345678901234567890", // 20 characters
			},
			wantErr: false,
		},
		{
			name: "empty login",
			creds: UserCredentials{
				Login:    "",
				Password: "12345678901234567890",
			},
			wantErr: true,
			errMsg:  "login is required",
		},
		{
			name: "empty password",
			creds: UserCredentials{
				Login:    "testuser",
				Password: "",
			},
			wantErr: true,
			errMsg:  "password is required",
		},
		{
			name: "password too short",
			creds: UserCredentials{
				Login:    "testuser",
				Password: "short",
			},
			wantErr: true,
			errMsg:  "password must be at least 8 characters",
		},
		{
			name: "password exactly 8 characters",
			creds: UserCredentials{
				Login:    "testuser",
				Password: "01234567",
			},
			wantErr: false,
		},
		{
			name: "password 7 characters",
			creds: UserCredentials{
				Login:    "testuser",
				Password: "0123456",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.creds.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
