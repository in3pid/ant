package ant

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"reflect"
)

// Query instantiates a cursor builder. The cursor ultimately reads
// sqlx.Rows from sqlx.DB.Query(q, rgs...) with either .Scan, .StructScan or .MapScan.
func Query(db *sqlx.DB, q string) TypeCurser {
	return dbQuery{db: db, q: q}
}

type dbQuery struct {
	db *sqlx.DB
	q  string
}

// Struct types cursor decoding to the given struct.
func (q dbQuery) Struct(s interface{}) Curser {
	t := reflect.TypeOf(s)
	switch s.(type) {
	case sql.Scanner:
		return scanRower{q, t}
	default:
		return structRower{q, t}
	}
}

// Map types cursor decoding to a map.
func (q dbQuery) Map() Curser {
	return mapRower{q}
}

// Cursor is the default-type cursor and is equal to Map()
func (q dbQuery) Cursor(args ...interface{}) Cursor {
	return q.Map().Cursor(args...)
}

/// a rowScanner wraps the (*sqlx.Rows).Scan call

type rowScanner interface {
	scanRow(row *sqlx.Rows) (T, error)
}

//- scanRower ses (*sql.Rows).Scan

func (r scanRower) Cursor(args ...interface{}) Cursor {
	return rows(r, r.q, args...)
}

func (r scanRower) scanRow(rows *sqlx.Rows) (T, error) {
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

func (r structRower) scanRow(rows *sqlx.Rows) (T, error) {
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

func (r mapRower) scanRow(rows *sqlx.Rows) (T, error) {
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
	return do(rowCursor{MakeSignal(), s, q, args})
}

func (c rowCursor) do() {
	var t T

	rows, err := c.Queryx(c.args...)
	for err == nil && rows.Next() {
		if t, err = c.scanRow(rows); err == nil {
			if !c.Send(t) {
				break
			}
		}
	}
	if err != nil {
		c.SendErr(err)
	}
	if rows != nil {
		rows.Close()
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
