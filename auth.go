package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type Key int

const MyKey Key = 0
const dataFile = "users.dat"

// Claims JWT schema of the data it will store
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Auth is used to pass the secret key to the validate and login handler
type Auth struct {
	secret   []byte
	users    map[string]User
	persist  func(Auth) error
	filepath string
}

func (a *Auth) validate(page http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie("Auth")
		if err != nil {
			log.Printf("No cookie: %v", err)
			http.Redirect(res, req, "/login", http.StatusTemporaryRedirect)
			return
		}

		token, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method")
			}
			return a.secret, nil
		})
		if err != nil {
			log.Printf("Didnt get the token")
			http.Redirect(res, req, "/login", http.StatusTemporaryRedirect)
			return
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			ctx := context.WithValue(req.Context(), MyKey, *claims)
			page.ServeHTTP(res, req.WithContext(ctx))
		} else {
			log.Printf("Dodgy claim")
			http.Redirect(res, req, "/login", http.StatusTemporaryRedirect)
			return
		}
	})
}
func (a *Auth) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "login", nil)
	} else if r.Method == "POST" {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")

		log.Printf("Logging in %v:%v", username, password)

		user := a.getUser(username)
		if user == nil {
			log.Println("Unknown user")
			renderTemplate(w, "login", "Unknown user")
		}
		err := bcrypt.CompareHashAndPassword(user.password, []byte(password))
		if err != nil {
			renderTemplate(w, "login", "Login Failed")
			return
		}

		expireToken := time.Now().Add(time.Hour * 1).Unix()
		expireCookie := time.Now().Add(time.Hour * 1)

		claims := Claims{
			"gawth",
			jwt.StandardClaims{
				ExpiresAt: expireToken,
				Issuer:    "localhost",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		signedToken, _ := token.SignedString(a.secret)

		cookie := http.Cookie{Name: "Auth", Value: signedToken, Expires: expireCookie, HttpOnly: true, Secure: true, Path: "/"}
		http.SetCookie(w, &cookie)
		log.Println("Logged In Ok")

		http.Redirect(w, r, "/", 307)

	} else {
		http.Error(w, "Invalid request method.", 405)
	}

}

func (a *Auth) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid method", 405)
	}

	u := NewUser(r.PostFormValue("username"), r.PostFormValue("password"))
	err := a.registerUser(u)

	if err != nil {
		renderTemplate(w, "login", err.Error())
	}
	http.Redirect(w, r, "/login/", 307)

}

func (a *Auth) registerUser(user User) error {
	if a.users == nil {
		a.users = make(map[string]User)
	}

	// Only allow 1 regsitered user.  Maybe one day I'll
	// support more but for now...nope
	if len(a.users) > 0 {
		return errors.New("Only one user allowed")
	}

	_, exists := a.users[user.username]
	if exists {
		return errors.New("User exists")
	}
	a.users[user.username] = user
	a.persist(*a)
	return nil

}

func (a *Auth) getUser(username string) *User {
	u := a.users[username]
	return &u
}

func persistUsers(a Auth) error {
	var buffer bytes.Buffer
	for _, val := range a.users {
		buffer.WriteString(fmt.Sprintf("%v,%v\n", val.username, val.password))
	}
	log.Printf(buffer.String())
	ioutil.WriteFile(a.filepath+dataFile, buffer.Bytes(), 0600)
	return nil
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid method", 405)
	}
	deleteCookie := http.Cookie{Name: "Auth", Value: "none", Expires: time.Now(), Secure: true, Path: "/"}
	http.SetCookie(w, &deleteCookie)
	http.Redirect(w, r, "/", 307)
}

// NewAuth - Create a new auth object - reads in user DB as well
func NewAuth(config Config, fn func(Auth) error) Auth {
	auth := Auth{secret: config.CookieKey, persist: fn, filepath: config.KeyLocation}
	auth.users = make(map[string]User)
	data, err := ioutil.ReadFile(config.KeyLocation + dataFile)
	if err == nil {
		log.Printf(string(data))
		lines := bytes.Split(data, []byte("\n"))
		log.Printf("Found %v\n ", len(lines))
		for _, line := range lines {
			log.Println("Looking")
			vals := bytes.Split(line, []byte(","))
			if len(vals) != 2 {
				log.Printf("Found %v in line\n ", len(vals))
				continue
			}

			user := NewUserWithHash(string(vals[0]), vals[1])

			auth.users[user.username] = user
		}

	}
	return auth
}

// User holds user data
type User struct {
	username string
	password []byte
}

// NewUser returns a user...will not store the password as plain text
func NewUser(username, password string) User {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	checkErr(err)
	return User{username: username, password: hash}
}

// NewUserWithHash returns a user...will not store the password as plain text
func NewUserWithHash(username string, password []byte) User {
	return User{username: username, password: password}
}
