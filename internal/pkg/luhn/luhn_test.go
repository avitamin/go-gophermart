package luhn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid luhn number",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "valid luhn number 2",
			number: "12345678903",
			want:   true,
		},
		{
			name:   "valid single digit",
			number: "0",
			want:   true,
		},
		{
			name:   "invalid luhn number",
			number: "79927398710",
			want:   false,
		},
		{
			name:   "empty string",
			number: "",
			want:   false,
		},
		{
			name:   "contains letters",
			number: "1234a5678",
			want:   false,
		},
		{
			name:   "contains special characters",
			number: "1234-5678",
			want:   false,
		},
		{
			name:   "valid credit card format",
			number: "4532015112830366",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Valid(tt.number)
			assert.Equal(t, tt.want, got)
		})
	}
}
