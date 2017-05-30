package main

import (
	"testing"
)

func Test(t *testing.T) {
	text := []byte("My name is Astaxie")
	key := []byte("the-key-has-to-be-32-bytes-long!")

	ciphertext, err := encrypt(text, key)
	if err != nil {
		t.Errorf("Failed to encrypt with err: %v", err)
	}

	plaintext, err := decrypt(ciphertext, key)
	if err != nil {
		t.Errorf("Failed to encrypt with err: %v", err)
	}
	if string(plaintext) != string(text) {
		t.Errorf("Expected : %v but got : %v", string(text), string(plaintext))
	}

}
