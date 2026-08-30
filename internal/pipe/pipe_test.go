package pipe_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uthereal/scheme-generator-go/internal/pipe"
)

// TestPipe_Do verifies that the Do method correctly chains multi-parameter
// functions to sequentially transform the internal value.
func TestPipe_Do(t *testing.T) {
	t.Run("Trim prefix and suffix", func(t *testing.T) {
		const input = "prefix_value_suffix"
		const expected = "value"

		p := pipe.NewPipe(input).
			Do(strings.TrimPrefix, "prefix_").
			Do(strings.TrimSuffix, "_suffix")

		assert.Equal(t, expected, p.Unwrap())
	})

	t.Run("Arithmetic operations", func(t *testing.T) {
		const input = 10
		const expected = 25

		add := func(a int, b int) int {
			return a + b
		}

		mul := func(a int, b int) int {
			return a * b
		}

		p := pipe.NewPipe(input).
			Do(add, 5).
			Do(mul, 2).
			Do(add, -5)

		assert.Equal(t, expected, p.Unwrap())
	})
}

// TestPipe_DoUnary verifies that the DoUnary method correctly chains unary
// functions to transform the internal value.
func TestPipe_DoUnary(t *testing.T) {
	t.Run("Unary string transformation", func(t *testing.T) {
		p := pipe.NewPipe("hello").
			DoUnary(strings.ToUpper)

		assert.Equal(t, "HELLO", p.Unwrap())
	})

	t.Run("Unary int transformation", func(t *testing.T) {
		double := func(v int) int {
			return v * 2
		}

		p := pipe.NewPipe(5).
			DoUnary(double)

		assert.Equal(t, 10, p.Unwrap())
	})
}
