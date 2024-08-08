package models

import (
	"database/sql"
	"os"
	"testing"
)

func newTestDB(t *testing.T) *sql.DB {
	//cfg := mysql.Config{
	//	User:            "test_web",
	//	Passwd:          "password",
	//	Addr:            "localhost:3306",
	//	DBName:          "test_snippetbox",
	//	ParseTime:       true,
	//	MultiStatements: true,
	//}
	//testDSN := flag.String("test-dsn", cfg.FormatDSN(), "MySQL Test DSN")
	//
	//t.Logf("flags parsed: %v", flag.Parsed())

	//db, err := sql.Open("mysql", *testDSN)
	db, err := sql.Open("mysql", "test_web:password@/test_snippetbox?parseTime=true&multiStatements=true")
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	t.Cleanup(func() {
		defer db.Close()

		script, err := os.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}

		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}
	})

	return db
}
