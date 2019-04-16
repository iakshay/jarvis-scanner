package common

import "testing"

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}

func TestIsAlive(t *testing.T) {
	ip := "172.217.10.14"
	s := IsAlive(ip)
	assertEqual(t, len(s), 1)
}
