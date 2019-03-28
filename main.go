package main

import (
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type certMapElm struct {
	CertHash string
	*x509.Certificate
}

type sessionRecord struct {
	ServerIP     string
	ServerPort   string
	ClientIP     string
	ClientPort   string
	TLSH         string
	Timestamp    time.Time
	JA3          string
	JA3Digest    string
	JA3S         string
	JA3SDigest   string
	Certificates []certMapElm
}

type sessions struct {
	Sessions []dbSession
}

type usessions struct {
	Sessions []sessionRecord
}

type dbSession struct {
	data string
	id   int
}

var db *sql.DB

func main() {
	// connect to db
	connStr := "user=postgres password=postgres dbname=passivessl"
	err := errors.New("")
	db, err = sql.Open("postgres", connStr)
	defer db.Close()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/index", indexHandler)
	http.HandleFunc("/ja3/", ja3Handler)
	http.HandleFunc("/ja3s/", ja3sHandler)

	err = http.ListenAndServe("localhost:8000", nil)
	if err != nil {
		log.Fatal("HTTP Server: ", err)
	}
}

func parseParam(req *http.Request, prefix string) (string, error) {
	param := strings.TrimPrefix(req.URL.Path, prefix)
	return param, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	s := usessions{}

	err := queryDB(&s, `
        SELECT
            id,
            data
        FROM sessions 
        ORDER BY id ASC`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(s)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, string(out))
}

func ja3Handler(w http.ResponseWriter, r *http.Request) {
	param, err := parseParam(r, "/ja3/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	s := usessions{}

	err = queryDB(&s, `
        SELECT
            id,
            data
		FROM sessions
		WHERE data ->> 'JA3Digest' = '`+param+`'
        ORDER BY id ASC	`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(s)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, string(out))
}

func ja3sHandler(w http.ResponseWriter, r *http.Request) {
	param, err := parseParam(r, "/ja3s/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	s := usessions{}

	err = queryDB(&s, `
        SELECT
            id,
            data
		FROM sessions
		WHERE data ->> 'JA3SDigest' = '`+param+`'
        ORDER BY id ASC	`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(s)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, string(out))
}

func queryDB(s *usessions, q string) error {
	rows, err := db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		sl := dbSession{}
		err = rows.Scan(
			&sl.id,
			&sl.data)
		if err != nil {
			return err
		}
		tmp := sessionRecord{}
		json.Unmarshal([]byte(sl.data), &tmp)
		s.Sessions = append(s.Sessions, tmp)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}
