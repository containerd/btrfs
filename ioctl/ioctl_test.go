package ioctl

import (
	"testing"
)

var casesIOC = []struct {
	got    uintptr
	expect uintptr
}{
	{got: IOC(1, 2, 3, 4), expect: 0x40040203},
}

func TestIOC(t *testing.T) {
	for i, c := range casesIOC {
		if c.got != c.expect {
			t.Errorf("unexpected ioc (case %d): %x(%b) vs %x(%b)",
				i+1, c.got, c.got, c.expect, c.expect)
		}
	}
}
