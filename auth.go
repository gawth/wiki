package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Key int

const MyKey Key = 0

// Claims JWT schema of the data it will store
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func validate(page http.Handler) http.Handler {
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
			return []byte("thisisasecret"), nil
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
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "login", nil)
	} else if r.Method == "POST" {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")

		log.Printf("Logging in %v:%v", username, password)

		if username == "gawth" && password == "fred" {
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

			signedToken, _ := token.SignedString([]byte("thisisasecret"))

			cookie := http.Cookie{Name: "Auth", Value: signedToken, Expires: expireCookie, HttpOnly: true, Secure: true, Path: "/"}
			http.SetCookie(w, &cookie)
			log.Println("Logged In Ok")

			http.Redirect(w, r, "/", 307)
		}

	} else {
		http.Error(w, "Invalid request method.", 405)
	}

}
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid method", 405)
	}
	deleteCookie := http.Cookie{Name: "Auth", Value: "none", Expires: time.Now(), Secure: true, Path: "/"}
	http.SetCookie(w, &deleteCookie)
	http.Redirect(w, r, "/", 307)
}
