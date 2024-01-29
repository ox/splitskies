package main

import (
	"fmt"
	"log/slog"
	mathrand "math/rand"
	"net/http"
	"splitskies/middleware"
	"splitskies/sms"
	"time"

	"splitskies/session"

	"github.com/gorilla/mux"
	"github.com/hako/durafmt"
	"github.com/jmoiron/sqlx"
)

type Session struct {
	User *User
	// TODO: track IP the session was created for to prevent session hijacking
	LoginPhone string
	Verified   bool
	Flash      string
}

type App struct {
	db   *sqlx.DB
	r    *mux.Router
	te   *TemplateEngine
	smss sms.SMSVerifier
	dir  string
	er   *ExpensesRepository
	ur   *UserRepository
	tr   *TripRepository
	ss   session.SessionStore[Session]
}

func (app *App) Init() {
	app.r = mux.NewRouter()
	app.ss = session.SessionStore[Session]{}

	app.ss.InitStore("SessionID", 14*24*time.Hour, "/login")

	app.ss.RawAddSession("adminsession", &Session{
		LoginPhone: "+19990008888",
		Verified:   true,
	})

	app.r.Use(middleware.Logging)

	app.r.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		})
	})

	app.r.HandleFunc("/", app.Index)
	app.r.HandleFunc("/login", app.LoginView).Methods("GET")
	app.r.HandleFunc("/login", app.Login).Methods("POST")
	app.r.HandleFunc("/login/check", app.LoginCheckView).Methods("GET")
	app.r.HandleFunc("/login/check", app.LoginCheck).Methods("POST")

	approuter := app.r.NewRoute().Subrouter()
	approuter.Use(app.ss.LoadSession)
	approuter.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s := app.ss.GetSessionFromCtx(r)
			// LoadSession should already redirect to login
			// if s == nil && r.URL.Path != "/login" {
			// 	http.Redirect(w, r, "/login", http.StatusFound)
			// 	return
			// }

			// If the phone isn't verified, go log in
			if !s.Verified {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			// If they're verified but don't have a user object yet
			if s.Verified && s.User == nil {
				if s.LoginPhone == "" {
					http.Redirect(w, r, "/login", http.StatusFound)
					return
				}

				u, ok, err := app.ur.GetUserByPhone(s.LoginPhone)
				if err != nil {
					slog.Error("could not get user by phone", "err", err)
					http.Error(w, "couldn't find you. try again later? k thnx", http.StatusInternalServerError)
					return
				}
				if ok {
					slog.Debug("got user by phone in middleware", "user", u)
					s.User = u
					app.ss.PutSession(w, r, s)
				}
				if !ok && r.URL.Path != "/register" {
					http.Redirect(w, r, "/register", http.StatusFound)
					return
				}
			}
			h.ServeHTTP(w, r)
		})
	})
	approuter.HandleFunc("/register", app.RegisterView).Methods("GET")
	approuter.HandleFunc("/register", app.Register).Methods("POST")
	approuter.HandleFunc("/app", app.Home).Methods("GET")
	approuter.HandleFunc("/app/trips/{tripid}", app.TripDetailsView).Methods("GET")
}

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir(app.dir))))
	app.r.ServeHTTP(w, r)
}

func (app *App) Index(w http.ResponseWriter, r *http.Request) {
	app.te.MustExecute(w, "index.html", nil)
}

func (app *App) ErrorView(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

type TripForDisplay struct {
	*Trip
	TimeSinceStart string
}

var welcomePhrases = []string{
	"Welcome home",
	"Finally, Splitskies",
	"Splitskies@Home",
	"It's better here — Splitskies",
	"Splitskies dreaming",
	"Veni, Vidi, Spliskies",
	"Eyyyy, howudoin?",
	"Howdy pardner",
	"Hi :)",
}

func genWelcomePhrase() string {
	return welcomePhrases[mathrand.Intn(len(welcomePhrases))]
}

func (app *App) Home(w http.ResponseWriter, r *http.Request) {
	s := app.ss.GetSessionFromCtx(r)
	tripsForUser, err := app.tr.GetTripsForUser(s.User.ID)
	if err != nil {
		slog.Error("could not get trip for user", "err", err, "userid", s.User.ID)
		app.ErrorView(w, r, fmt.Errorf("could not get your trips :("))
		return
	}

	tripsToShow := make([]*TripForDisplay, 0)
	for _, trip := range tripsForUser {
		timeSinceCreated := durafmt.ParseShort(time.Since(trip.CreatedAt))
		tripsToShow = append(tripsToShow, &TripForDisplay{trip, timeSinceCreated.String() + " ago"})
	}

	data := struct {
		Username      string
		Flash         string
		Trips         []*TripForDisplay
		WelcomePhrase string
	}{
		s.User.Username,
		s.Flash,
		tripsToShow,
		genWelcomePhrase(),
	}

	app.te.MustExecute(w, "app.html", data)
}

type TripDetailsExpenseForShow struct {
	*Expense
	Date string
	Cost string
}

func (app *App) TripDetailsView(w http.ResponseWriter, r *http.Request) {
	s := app.ss.GetSessionFromCtx(r)
	vars := mux.Vars(r)
	trip, err := app.tr.GetTripForUser(vars["tripid"], s.User.ID)
	if err != nil {
		slog.Error("could not get trip for user", "err", err, "userid", s.User.ID, "tripid", vars["tripid"])
		app.ErrorView(w, r, fmt.Errorf("could not get your trip :("))
		return
	}

	expenses, err := app.er.GetTripExpenses(trip.ID)
	if err != nil {
		slog.Error("could not get trip expenses", "err", err)
		app.ErrorView(w, r, fmt.Errorf("whoops! coudn't get your trip expenses. try again later maybe?"))
		return
	}

	tripExpensesForShow := make([]*TripDetailsExpenseForShow, len(expenses))
	for i, expense := range expenses {
		tripExpensesForShow[i] = &TripDetailsExpenseForShow{
			expense,
			expense.CreatedAt.Format("Jan 2, 3:04 PM"),
			fmt.Sprintf("$%d.%02d", expense.OwnerCostCents/100, expense.OwnerCostCents%100),
		}
	}

	data := struct {
		Username string
		Trip     *Trip
		Expenses []*TripDetailsExpenseForShow
	}{
		s.User.Username,
		trip,
		tripExpensesForShow,
	}

	app.te.MustExecute(w, "trip_detail.html", data)
}
