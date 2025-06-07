package main

import "testing"

func TestNewClient(t *testing.T) {
	c, plain_key, err := NewClient("the test client")
	if err != nil {
		t.Fatalf(err.Error())
	}
	if plain_key == "" || plain_key == c.Key {
		t.Fatalf("Bad plan_key %s %s", plain_key, c.Key)
	}
}

func TestClientUpdateKey(t *testing.T) {
	c, plain_key, err := NewClient("the test client")
	if err != nil {
		t.Fatalf(err.Error())
	}

	new_key, err := c.UpdateKey()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if plain_key == new_key {
		t.Fatalf("key not updated %s %s", plain_key, new_key)
	}

}
