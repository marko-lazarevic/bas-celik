package card

import (
	"testing"

	"github.com/ebfe/scard"
)

func TestFormatState(t *testing.T) {
	tests := []struct {
		input    scard.StateFlag
		expected string
	}{
		{scard.StateUnaware, "StateUnaware"},
		{scard.StateIgnore, "StateIgnore"},
		{scard.StateChanged, "StateChanged"},
		{scard.StateUnknown, "StateUnknown"},
		{scard.StatePresent, "StatePresent"},
		{scard.StateAtrmatch, "StateAtrmatch"},
		{scard.StateExclusive, "StateExclusive"},
		{scard.StateMute, "StateMute"},
		{scard.StateUnpowered, "StateUnpowered"},
		{scard.StatePresent | scard.StateChanged, "StateChanged StatePresent"},
	}

	for _, test := range tests {
		result := FormatState(test.input)
		if result != test.expected {
			t.Errorf("FormatState(%v) = %q; want %q", test.input, result, test.expected)
		}
	}
}