package db

import (
	"joynext/downdetector/internal/utils"

	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gorilla/sessions"
)

// UserDB contains user username and password.
type UserDB struct {
	Username string `db:"username"`
	Password string `db:"password"`
}

// UserJSON represents a user in JSON format.
type UserJSON struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Pepper   string `json:"pepper"` //
}

var store = sessions.NewCookieStore([]byte(utils.GenerateRandomString(20)))

// @LoginMiddleware authenticates a user using the provided login credentials.
//
// @Summary Authenticate user
// @Description Authenticate a user using their login credentials.
// @Description Password: sha256(sha256(password + salt) + pepper)
// @Tags user
// @Accept json
// @Produce plain
// @Param user body UserJSON true "User credentials"
// @Success 200
// @Failure 403
// @Failure 500
// @Router /login [post]
func LoginMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := unmarshallUser(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error("Failed to marshall user", "err", err)
			return
		}

		// Add user to request context
		ctx := r.Context()
		req := r.WithContext(context.WithValue(ctx, "user", user))
		*r = *req

		res := DB.QueryRow("SELECT password FROM users WHERE username=?", user.Username)
		var hash string

		err = res.Scan(&hash)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error("Failed to select user from DB", "err", err)
			return
		}

		h := sha256.New()
		h.Write([]byte(hash + user.Pepper))
		hashedWithPepper := hex.EncodeToString(h.Sum(nil))

		// Check if the provided password + salt matches the stored hash + salt
		if string(hashedWithPepper) != user.Password {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		f(w, r)
	}
}

// SessionHandler creates a session for an authenticated user and redirects them to the referrer URL.
func SessionHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Failed to get session", "err", err)
		return
	}
	session.Options.MaxAge = 60 * 60 // cookie valid for 1 hour

	session.Values["authenticated"] = true
	session.Values["username"] = r.Context().Value("user").(UserJSON).Username

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		log.Error("Failed to save to session", "err", err)
		return
	}

	ip := r.RemoteAddr
	utils.NoReportLog.Infof("New login from %s", ip)

	// Get location from URL, e.g., "/dashboard" from "localhost?ref=/dashboard"
	referer := strings.Split(strings.Split(r.Referer(), "?")[1], "=")[1]

	w.Header().Add("Location", referer)
	w.WriteHeader(http.StatusOK)
}

// CheckIfUserLoggedIn middleware checks if a user is logged in.
func CheckIfUserLoggedIn(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "auth")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error("Failed to get session", "err", err)
			return
		}

		authenticated, ok := session.Values["authenticated"].(bool)
		if !ok || !authenticated {
			url := fmt.Sprintf("/login?ref=%s", r.URL)
			http.Redirect(w, r, url, http.StatusSeeOther) // has to be 3XX, so the browser will automatically redirect
			return
		}

		// User is authenticated, proceed with displaying profile
		_, ok = session.Values["username"].(string)
		if !ok {
			http.Error(w, "Username not found in session", http.StatusInternalServerError)
			log.Error("User not found in session", "err", err)
			return
		}
		f(w, r)
	}
}

// @LogoutHandler logs the user out by invalidating the session.
//
// @Summary Log out user
// @Description Log out the currently authenticated user by invalidating the session.
// @Tags user
// @Produce plain
// @Success 303
// @Failure 500
// @Router /logout [get]
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Failed to get session", "err", err)
		return
	}

	// Invalidate current session cookie
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Failed to save to session", "err", err)
		return
	}

	ip := r.RemoteAddr
	utils.NoReportLog.Infof("Logged out from %s", ip)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// @ChangePasswordHandler allows a user to change their password.
//
// @Summary Change user password
// @Description Allows an authenticated user to change their password.
// @Tags user
// @Accept json
// @Produce plain
// @Success 303
// @Failure 500
// @Router /change-password [post]
func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Failed to get session", "err", err)
		return
	}

	user, err := unmarshallUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Failed to unmarshall user", "err", err)
		return
	}

	username := session.Values["username"].(string)
	newPassword := user.Password

	_, err = DB.Exec("UPDATE users set password=? WHERE username=?", newPassword, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Failed to marshall user", "err", err)
		return
	}

	ip := r.RemoteAddr
	utils.NoReportLog.Infof("%s changed password", ip)

	w.Header().Add("Location", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

// @GetSaltHandler generates and returns a salt.
//
// @Summary Get user salt
// @Description Generates and returns a salt assigned to the user.
// @Tags user
// @Produce plain
// @Success 200 {string} body salt
// @Router /salt [get]
func GetSaltHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")
	if username == "" {
		session, err := store.Get(r, "auth")
		authenticated, ok := session.Values["authenticated"].(bool)

		if !ok || !authenticated || err != nil {
			http.Error(w, "No username provided", http.StatusForbidden)
			return
		}
		username = session.Values["username"].(string)
	}
	res := DB.QueryRow("SELECT salt FROM users WHERE username=?", username)

	var salt string

	err := res.Scan(&salt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(fmt.Sprintf("Failed to select salt for user %s", username), "err", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(salt))
}

// @GetPepperHandler generates and returns a salt.
//
// @Summary Get one time salt
// @Description Generates and returns a one time salt used to login.
// @Tags user
// @Produce plain
// @Success 200 {string} body pepper
// @Router /pepper [get]
func GetPepperHandler(w http.ResponseWriter, r *http.Request) {
	pepper := utils.GenerateRandomString(5)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pepper))
}

// unmarshallUser reads the request body and unmarshals it into a UserJSON string.
func unmarshallUser(r *http.Request) (UserJSON, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return UserJSON{}, err
	}

	user := UserJSON{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return UserJSON{}, err
	}

	return user, nil
}
