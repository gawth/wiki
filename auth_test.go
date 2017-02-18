package main

import "testing"
import "golang.org/x/crypto/bcrypt"

const user = "fred"
const pword = "12345"

func stubPersist(a Auth) error {
	return nil
}

func TestNewUser(t *testing.T) {
	actual := NewUser(user, pword)

	if string(actual.password) == pword {
		t.Errorf("TestNewUser: User returned has a plain text password :-(")
	}

	validPassword := bcrypt.CompareHashAndPassword(actual.password, []byte(pword))
	if validPassword != nil {
		t.Errorf("TestNewUser: Password doesnt check out")
	}
}

func TestRegisterUser(t *testing.T) {
	target := Auth{persist: stubPersist}
	data := User{username: user, password: []byte(pword)}

	err := target.registerUser(data)

	if err != nil {
		t.Errorf("TestRegisterUser: Failed to register new user")
	}

	if target.getUser(user) == nil {
		t.Errorf("TestRegisterUser: Failed to retrieve user")
	}
}
func TestDuplicateRegisterUser(t *testing.T) {
	target := Auth{persist: stubPersist}
	data := User{username: user, password: []byte(pword)}

	err := target.registerUser(data)

	if err != nil {
		t.Errorf("TestDuplicateRegisterUser: Failed to register new user")
	}

	err = target.registerUser(data)

	if err == nil {
		t.Errorf("TestDuplicateRegisterUser: Registered a duplicate user")
	}

}
func TestRegisterOnlyOneUser(t *testing.T) {
	target := Auth{persist: stubPersist}
	data := User{username: user, password: []byte(pword)}

	err := target.registerUser(data)

	if err != nil {
		t.Errorf("TestRegisterOnlyOneUser: Failed to register new user")
	}

	user2 := NewUser("fred2", "klshdf98")
	err = target.registerUser(user2)

	if err == nil {
		t.Errorf("TestRegisterOnlyOneUser: Managed to register more than one user")
	}

}
