package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"splitskies/config"
	"splitskies/sms"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lmittmann/tint"
	_ "modernc.org/sqlite"
)

func init() {
	// set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
}

func main() {
	configPath := flag.String("config", "", "Path to app config")
	flag.Parse()

	if configPath == nil || *configPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	conf, err := config.FromFile(*configPath)
	if err != nil {
		log.Fatalf("could not parse config: %s", err)
	}

	db, err := sqlx.Open("sqlite", conf.DBConfig.DBPath)
	if err != nil {
		fmt.Printf("could not open db at %s: %s\n", conf.DBConfig.DBPath, err)
		os.Exit(1)
	}
	defer db.Close()

	te, err := NewTemplateEngine()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	userRepo := &UserRepository{db: db}
	if err := userRepo.CreateTable(); err != nil {
		slog.Error("could not create users table: %w", err)
		os.Exit(1)
	}

	tripRepo := &TripRepository{db: db}
	if err := tripRepo.CreateTable(); err != nil {
		slog.Error("could not create trips table: %w", err)
		os.Exit(1)
	}

	expensesRepo := &ExpensesRepository{db: db}
	if err := expensesRepo.CreateTable(); err != nil {
		slog.Error("could not create expenses table: %w", err)
		os.Exit(1)
	}

	smss := sms.Connect(conf.TwilioConfig)
	app := &App{
		db:   db,
		te:   te,
		smss: smss,
		dir:  conf.AppConfig.PublicPath,
		er:   expensesRepo,
		ur:   userRepo,
		tr:   tripRepo,
	}
	app.Init()

	httpmux := http.NewServeMux()
	httpmux.Handle("/", app)

	ctx := context.Background()
	addrStr := conf.AppConfig.Bind
	server := &http.Server{
		Addr:    addrStr,
		Handler: httpmux,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	fmt.Printf("Listening on %s", addrStr)
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
	}
}
