package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"martinop.net/snippetbox/pkg/models/mysql"

	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *mysql.SnippetModel
	templateCache map[string]*template.Template
}

func main() {
	//command-line flag "addr", default value of :4000, allows runtime edit
	//ports 0-1023 are restricted and can only be used by services with root privileges
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:password@/snippetbox?parseTime=true", "MySQL database")
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

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		snippets:      &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
	}

	//http.Server struct
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr) //log.Println("Starting server on %s", *addr)
	error := srv.ListenAndServe()                  //err := http.ListenAndServe(*addr, mux)
	errorLog.Fatal(error)                          //instead of log.Fatal(err)
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
