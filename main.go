package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type sessions struct {
	Sessions []dbSession
}

type dbSession struct {
	data string
	id   int
}

var db *sql.DB

func main() {

	http.HandleFunc("/index", indexHandler)

	err := http.ListenAndServe("localhost:8000", nil)
	if err != nil {
		log.Fatal("HTTP Sever: ", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	s := sessions{}

	err := querySessions(&s)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, s.Sessions[0].data)
	// for _, s := range s.Sessions {
	// 	fmt.Fprintf(w, s.data)
	// }
}

func querySessions(s *sessions) error {
	// connect to db // PARKING HERE
	connStr := "user=postgres password=postgres dbname=passivessl"
	db, err := sql.Open("postgres", connStr)
	defer db.Close()
	if err != nil {
		panic(err)
	}

	rows, err := db.Query(`
        SELECT
            id,
            data
        FROM sessions 
        ORDER BY id ASC`)
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
		s.Sessions = append(s.Sessions, sl)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}
