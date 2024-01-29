package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (app *App) LoginView(w http.ResponseWriter, r *http.Request) {
	s := app.ss.GetSessionFromRequest(r)
	if s == nil {
		slog.Debug("no existing session")
		s = &Session{
			Verified: false,
		}
		app.ss.PutSession(w, r, s)
	}

	if s.Verified {
		http.Redirect(w, r, "/app", http.StatusFound)
		return
	}

	app.te.MustExecute(w, "login.html", nil)
}

func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Printf("could not parse form: %s\n", err)
		return
	}

	phone := r.Form.Get("phone")
	if phone == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please enter a phone number"))
		return
	}

	// Add US country code
	phone = "+1" + phone

	s := app.ss.GetSessionFromRequest(r)
	if s == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	s.LoginPhone = phone

	if err := app.smss.Send(phone); err != nil {
		slog.Error("could not send verification", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not send a verification request, try again later"))
		return
	}

	app.ss.PutSession(w, r, s)
	http.Redirect(w, r, "/login/check", http.StatusFound)
}

func (app *App) LoginCheckView(w http.ResponseWriter, r *http.Request) {
	s := app.ss.GetSessionFromRequest(r)
	if s == nil || s.LoginPhone == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	data := struct {
		Flash string
	}{
		Flash: s.Flash,
	}

	app.te.MustExecute(w, "login_check.html", data)
}

func (app *App) LoginCheck(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Printf("could not parse form: %s\n", err)
		return
	}

	code := r.Form.Get("code")
	s := app.ss.GetSessionFromRequest(r)
	if s == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	phone := s.LoginPhone
	approved, err := app.smss.Check(phone, code)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("could not check code: %s", err.Error())))
		return
	}

	if approved {
		s.Verified = true

		u, ok, err := app.ur.GetUserByPhone(phone)
		if !ok && err == nil {
			// If there's no user with this phone number, let them register
			app.ss.PutSession(w, r, s)
			http.Redirect(w, r, "/register", http.StatusFound)
			return
		}

		if err != nil {
			slog.Error("could not get user by phone: %w", err)
			http.Error(w, "there was a server error, try again later", http.StatusInternalServerError)
			return
		}
		s.User = u
		app.ss.PutSession(w, r, s)
		http.Redirect(w, r, "/app", http.StatusFound)
		return
	} else {
		s.Flash = "Invalid code!"
		app.ss.PutSession(w, r, s)
		http.Redirect(w, r, "/login/check", http.StatusSeeOther)
	}
}

func (app *App) RegisterView(w http.ResponseWriter, r *http.Request) {
	s := app.ss.GetSessionFromCtx(r)
	if s == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	_, ok, err := app.ur.GetUserByPhone(s.LoginPhone)
	if err != nil {
		slog.Error("could not get user by phone", "err", err)
		return
	}
	if ok {
		app.ss.DeleteSession(r)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	app.te.MustExecute(w, "register.html", nil)
}

func (app *App) Register(w http.ResponseWriter, r *http.Request) {
	s := app.ss.GetSessionFromCtx(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "couldn't parse the form", http.StatusBadRequest)
		return
	}

	username := r.Form.Get("username")
	if username == "" {
		http.Error(w, "missing username", http.StatusBadRequest)
		return
	}

	u, err := app.ur.AddUser(username, s.LoginPhone)
	if err != nil {
		slog.Error("could not register new user", "err", err, "username", username)
		http.Error(w, "couldn't create that user", http.StatusInternalServerError)
	}

	s.User = u
	app.ss.PutSession(w, r, s)
	http.Redirect(w, r, "/app", http.StatusFound)
}
