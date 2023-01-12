package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"martinop.net/snippetbox/pkg/models/mysql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golangcollege/sessions"
)

type contextKey string

var contextKeyUser = contextKey("user")

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	session       *sessions.Session
	snippets      *mysql.SnippetModel
	templateCache map[string]*template.Template
	users         *mysql.UserModel
}

func main() {
	//command-line flag "addr", default value of :4000, allows runtime edit
	//ports 0-1023 are restricted and can only be used by services with root privileges
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:password@/snippetbox?parseTime=true", "MySQL database")
	//command-line flag to authenticate session cookies
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")
	//needs to be added BEFORE using the addr variable, otherwise it will always contain the default value
	flag.Parse()

	//destination to write logs, prefix for messages, flags with additional information
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile) //log.Llongfile for full path, log.LUTC for UTC time

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close() //close the db pool when terminating app

	templateCache, err := newTemplateCache("./ui/html")
	if err != nil {
		errorLog.Fatal(err)
	}

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
		snippets:      &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
		users:         &mysql.UserModel{DB: db},
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,                                     // Go's favoured cipher suites given preference - strong cipher suite / forward secrecy
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256}, // less CPU intensive than other options
	}

	//http.Server struct
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,      // if ReadTimeOut is set but IdleTimeout isn't, IdleTimeout defaults to ReadTimeout
		ReadTimeout:  5 * time.Second,  // short read timeout = prevents slow client attacks
		WriteTimeout: 10 * time.Second, //
	}

	infoLog.Printf("Starting server on %s", *addr)
	error := srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem") // both files in gitignore
	errorLog.Fatal(error)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
