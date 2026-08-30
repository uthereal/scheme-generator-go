package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uthereal/scheme-generator-go/internal/ast"
)

func TestEnum_ToGoPascalName(t *testing.T) {
	tests := []struct {
		name string
		enum ast.Enum
		want string
	}{
		{
			name: "snake case enum",
			enum: ast.Enum{
				Name:   "order_status",
				Values: []string{"PENDING", "COMPLETED"},
			},
			want: "OrderStatus",
		},
		{
			name: "single word enum",
			enum: ast.Enum{
				Name:   "status",
				Values: []string{"ACTIVE", "INACTIVE"},
			},
			want: "Status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.enum.ToGoPascalName())
		})
	}
}

func TestEnum_ValueToGoPascalName(t *testing.T) {
	e := ast.Enum{}

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "uppercase active",
			value: "ACTIVE",
			want:  "Active",
		},
		{
			name:  "snake case uppercase",
			value: "IN_PROGRESS",
			want:  "InProgress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, e.ValueToGoPascalName(tt.value))
		})
	}
}
