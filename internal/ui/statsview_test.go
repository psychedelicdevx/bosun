package ui

import "testing"

func TestHumanBytes(t *testing.T) {
	cases := map[uint64]string{
		512:        "512B",
		1024:       "1.0KiB",
		1536:       "1.5KiB",
		1048576:    "1.0MiB",
		24788992:   "23.6MiB",
		1073741824: "1.0GiB",
	}
	for in, want := range cases {
		if got := humanBytes(in); got != want {
			t.Errorf("humanBytes(%d) = %q, want %q", in, got, want)
		}
	}
}
