package ant

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"reflect"
)

//
type Blob map[string]interface{}

// Query instantiates a cursor builder. The resulting Curser.Cursor(args...) ultimately reads
// sqlx.Rows from sqlx.DB.Query(q, args...) with either .Scan, .StructScan or .MapScan.
func QueryDB(db *sqlx.DB, q string) TypeCurser {
	return dbQuery{db: db, q: q}
}

type dbQuery struct {
	db *sqlx.DB
	q  string
}

// Type specifies the cursor to decode to the given type.
func (q dbQuery) Type(t interface{}) Curser {
	tt := reflect.TypeOf(t)
	for tt.Kind() == reflect.Ptr {
		tt = tt.Elem()
	}

	switch t.(type) {
	case sql.Scanner:
		return scanRower{q, tt}
	case Blob:
		return mapRower{q}
	default:
		return structRower{q, tt}
	}
}

// Cursor defaults to a common Blob type.
func (q dbQuery) Cursor(args ...interface{}) Cursor {
	return q.Type((Blob)(nil)).Cursor(args...)
}

/// a rowScanner wraps the (*sqlx.Rows).Scan call

type rowScanner interface {
	scanRow(row *sqlx.Rows) (Value, error)
}

//- scanRower uses (*sql.Rows).Scan

func (r scanRower) Cursor(args ...interface{}) Cursor {
	return rows(r, r.q, args...)
}

func (r scanRower) scanRow(rows *sqlx.Rows) (Value, error) {
	p := reflect.New(r.t).Interface()
	e := rows.Scan(p)
	return p, e
}

type scanRower struct {
	q dbQuery
	t reflect.Type
}

//- structRower uses (*sqlx.Rows).StructScan

func (r structRower) Cursor(args ...interface{}) Cursor {
	return rows(r, r.q, args...)
}

func (r structRower) scanRow(rows *sqlx.Rows) (Value, error) {
	p := reflect.New(r.t).Interface()
	e := rows.StructScan(p)
	return p, e
}

type structRower struct {
	q dbQuery
	t reflect.Type
}

//-- mapRowwer uses (*sqlx.Rows).MapScan

func (r mapRower) Cursor(args ...interface{}) Cursor {
	return rows(r, r.q, args...)
}

func (r mapRower) scanRow(rows *sqlx.Rows) (Value, error) {
	m := make(map[string]interface{})
	e := rows.MapScan(m)
	return m, e
}

type mapRower struct{ q dbQuery }

//-- row-copying cursor

type rowCursor struct {
	Signal
	rowScanner
	dbQuery
	args []interface{}
}

func rows(s rowScanner, q dbQuery, args ...interface{}) Cursor {
	return Do(rowCursor{NewSignal(), s, q, args})
}

func (c rowCursor) Do() {
	rows, err := c.Queryx(c.args...)
	if err != nil {
		panic(err)
	}
	defer func() {
		// mysql driver wont close until it has read to eof
		go rows.Close()
	}()

	for rows.Next() {
		t, err := c.scanRow(rows)
		if err != nil {
			panic(err)
		}
		if !Send(c, t) {
			break
		}
	}
}

//--

func (q dbQuery) Exec(args ...interface{}) (sql.Result, error) {
	return q.db.Exec(q.q, args...)
}

func (q dbQuery) Queryx(args ...interface{}) (*sqlx.Rows, error) {
	return q.db.Queryx(q.q, args...)
}

func (q dbQuery) QueryRowx(args ...interface{}) *sqlx.Row {
	return q.db.QueryRowx(q.q, args...)
}
