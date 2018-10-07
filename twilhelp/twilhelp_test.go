package twilhelp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanNumber(t *testing.T) {
	type cleanTest struct {
		orig, cleaned string
	}

	cleanTests := []cleanTest{
		{"+11234567890", "+11234567890"},
		{"123-456-1987", "+11234561987"},
		{"123 456 1987", "+11234561987"},
		{"1234561987", "+11234561987"},
		{"1 123-456-1987", "+11234561987"},
		{"+1 (123) 321-1234", "+11233211234"},
		{"+44 7911 123456", "+447911123456"},
	}

	for _, tt := range cleanTests {
		got := cleanNumber(tt.orig)
		assert.Equal(t, tt.cleaned, got)
	}
}
