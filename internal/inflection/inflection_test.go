package inflection_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uthereal/scheme-generator-go/internal/inflection"
)

func TestSingular(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "nexus stays nexus",
			input: "nexus",
			want:  "nexus",
		},
		{
			name:  "nexuses becomes nexus",
			input: "nexuses",
			want:  "nexus",
		},
		{
			name:  "campus stays campus",
			input: "campus",
			want:  "campus",
		},
		{
			name:  "bonus stays bonus",
			input: "bonus",
			want:  "bonus",
		},
		{
			name:  "canvas stays canvas",
			input: "canvas",
			want:  "canvas",
		},
		{
			name:  "lens stays lens",
			input: "lens",
			want:  "lens",
		},
		{
			name:  "species stays species",
			input: "species",
			want:  "species",
		},
		{
			name:  "users becomes user",
			input: "users",
			want:  "user",
		},
		{
			name:  "accounts becomes account",
			input: "accounts",
			want:  "account",
		},
		{
			name:  "status stays status",
			input: "status",
			want:  "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, inflection.Singular(tt.input))
		})
	}
}

func TestPlural(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "nexus becomes nexuses",
			input: "nexus",
			want:  "nexuses",
		},
		{
			name:  "user becomes users",
			input: "user",
			want:  "users",
		},
		{
			name:  "campus becomes campuses",
			input: "campus",
			want:  "campuses",
		},
		{
			name:  "species stays species",
			input: "species",
			want:  "species",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, inflection.Plural(tt.input))
		})
	}
}
