package docker

import "testing"

func TestNewWithHostSSHConstructs(t *testing.T) {
	// ssh:// is lazy: the client constructs and the ssh connection is only
	// dialed on the first API call, so construction must succeed offline.
	c, err := NewWithHost("ssh://user@example.com")
	if err != nil {
		t.Fatalf("ssh host should construct a client, got %v", err)
	}
	if c == nil || c.cli == nil {
		t.Fatal("ssh host should return a usable client")
	}
}

func TestNewWithHostTCPConstructs(t *testing.T) {
	c, err := NewWithHost("tcp://10.0.0.5:2375")
	if err != nil {
		t.Fatalf("tcp host should construct a client, got %v", err)
	}
	if c == nil || c.cli == nil {
		t.Fatal("tcp host should return a usable client")
	}
}
