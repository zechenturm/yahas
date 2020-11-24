package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
	"time"
	"html/template"
	"encoding/base64"
	"io/ioutil"
)

type userName string
type sessionid string

var cookieJar map[userName]sessionid

func createWebserver(useLogins bool) *mux.Router {
	cookieJar = make(map[userName]sessionid)
	router = mux.NewRouter()

	if useLogins {
		router.Use(middleware)
		router.HandleFunc("/login", loginHandler)
		router.HandleFunc("/start", startHandler)
	}

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/"))))
	return router
}


func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("intercepted request for:", r.URL)
		if r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
		} else {
			if isLoggedIn(r) {
				fmt.Println("user logged in")
				next.ServeHTTP(w, r)
			} else {
				fmt.Println("user not logged in")
				if data := r.Header.Get("Authorization"); data != "" {
					fmt.Println(data)
					data = data[len("Basic "):]
					decoded := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
					bytes, err := ioutil.ReadAll(decoded)
					if err != nil {
						fmt.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					data = string(bytes)
					split := strings.Index(data, ":")
					user := data[:split]
					password := data[split+1:]

					users, err := readUsers()
					if err != nil {
						fmt.Println("error reading users:", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					if !checkUser(users, user, password) {
						fmt.Println(user, "failed to login")
						w.WriteHeader(http.StatusUnauthorized)
						w.Write([]byte("wrong username/password!\n"))
						return
					}
					fmt.Println(user, "logged in successfully")
					cookie := makeCookie(user)
					http.SetCookie(w, &cookie)
					fmt.Println("calling next")
					next.ServeHTTP(w, r)

				} else {
					w.Header().Add("Cache-Control", "no-store")
					w.Header().Add("WWW-Authenticate", "Basic realm=\"hello\"")
					w.WriteHeader(http.StatusUnauthorized)
				}
			}
		}
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fmt.Println("got login  request:")
	fmt.Println("Method:", r.Method)
	if r.Method == "GET" {
		if err := sendLoginSite(w); err != nil {
			fmt.Println("Error opening login form:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		user, authenticated, err := checkCredentials(r)
		if err != nil {
			fmt.Println("Error opening login form:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !authenticated {
			sendLoginFailure(w, user)
			return
		}
		fmt.Println(user, "logged in successfully")
		redirectWithCookie(w, r, user)
	}
}

func sendLoginSite(w http.ResponseWriter) error {
	t, err := template.ParseFiles("./web/loginform.html")
	if err != nil {
		return err
	}

	err = t.Execute(w, "")
	if err != nil {
		return err
	}
	return nil
}

func redirectWithCookie(w http.ResponseWriter, r *http.Request, user string) {


	redirectURL := "/"
	redirectCookie, err := r.Cookie("SessionID")
	if err == nil {
		redirectURL = redirectCookie.Value
	}

	cookie := makeCookie(user)
	http.SetCookie(w, &cookie)

	fmt.Println("redirecting to", redirectURL)

	http.Redirect(w, r, redirectURL, 301)
}

func checkCredentials(r *http.Request) (string, bool, error) {
	r.ParseForm()
	user := r.Form.Get("user")
	password := r.Form.Get("password")

	users, err := readUsers()
	if err != nil {
		return user, false, err
	}
	return user, checkUser(users, user, password), nil
}

func sendLoginFailure(w http.ResponseWriter, user string) {
	fmt.Println(user, " failed to log in")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("wrong username/password!\n"))
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "start!")
}

func makeCookie(user string) http.Cookie {
	hasher := sha256.New()
	randomBytes := make([]byte, 128)
	rand.Read(randomBytes)
	hasher.Write(randomBytes)
	hash := hasher.Sum(nil)
	value := user + ":" + hex.EncodeToString(hash)
	cookieJar[userName(user)] = sessionid(value)
	expire := time.Now().AddDate(0, 0, 1)
	return http.Cookie{Name: "SessionID", Value: value, Expires: expire, HttpOnly: true}
}

func isLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("SessionID")
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Println(cookie)
	fields := strings.Split(cookie.Value, ":")
	if len(fields) != 2 {
		fmt.Println("error parsing cookie fields: wrong number of fields: ", len(fields))
		return false
	}
	expectedSessionID, ok := cookieJar[userName(fields[0])]
	if !ok {
		return false
	}
	return sessionid(cookie.Value) == expectedSessionID
}
