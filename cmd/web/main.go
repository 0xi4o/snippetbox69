package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/charmbracelet/log"
	"github.com/go-playground/form/v4"
	"github.com/go-sql-driver/mysql"
	"html/template"
	"net/http"
	"os"
	"snippetbox.i4o.dev/internal/models"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	logger         *log.Logger
	snippets       *models.SnippetModel
	templates      map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	// define a varialbe to capture command line argument
	addr := flag.String("addr", ":4000", "HTTP Network Address")
	// create a config for the dsn
	cfg := mysql.Config{
		User:      "web",
		Passwd:    "snippetbox",
		Addr:      "localhost:3306",
		DBName:    "snippetbox",
		ParseTime: true,
	}
	dsn := flag.String("dsn", cfg.FormatDSN(), "MySQL DSN")
	flag.Parse()

	// create a structured logger
	//logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	//logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
	})

	// open a connection pool to the database
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	// when main() exits, close all connections
	defer db.Close()

	templates, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	// add snippets model with connection pool to app dependencies
	app := &application{
		formDecoder:    formDecoder,
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		templates:      templates,
		sessionManager: sessionManager,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     logger.StandardLog(log.StandardLogOptions{ForceLevel: log.ErrorLevel}),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// log the starting server
	//log.Printf("Starting server on %s", *addr)
	logger.Info("Starting server:", "addr", *addr)

	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

	// log if there are errors and exit
	//log.Fatalln(err)
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
