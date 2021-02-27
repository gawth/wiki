package main

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestSqlStuff(t *testing.T) {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		t.Error(err)
	}

	_, err = db.Exec(`DROP TABLE test`)
	if err != nil {
		// Log it but carry on...probably doesn't exist...
		t.Log(err)
	}

	_, err = db.Exec(`CREATE TABLE test (
  	uid INTEGER PRIMARY KEY AUTOINCREMENT,
  	username VARCHAR(64) NULL,
  	departname VARCHAR(64) NULL,
  	created DATETIME)`)
	if err != nil {
		t.Error(err)
	}

	stmt, err := db.Prepare("INSERT INTO test(username, departname, created) values(?,?,?)")
	if err != nil {
		t.Error(err)
	}

	_, err = stmt.Exec("gawth", "home", "2021-02-13")
	if err != nil {
		t.Error(err)
	}

	rows, err := db.Query("SELECT * FROM test")
	if err != nil {
		t.Error(err)
	}
	var uid int
	var username string
	var department string
	var created time.Time

	for rows.Next() {
		err = rows.Scan(&uid, &username, &department, &created)
		checkErr(err)
		fmt.Println(uid)
		fmt.Println(username)
		fmt.Println(department)
		fmt.Println(created)
	}

	rows.Close()

	db.Close()
}
