package ant

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"reflect"
)

///

func QueryDB(db *sqlx.DB, q string) TypeCurser {
	return dbQuery{db: db, q: q}
}

type dbQuery struct {
	db *sqlx.DB
	q  string
}

func (q dbQuery) Struct(s interface{}) Curser {
	t := reflect.TypeOf(s)
	switch s.(type) {
	case sql.Scanner:
		return scanRower{q, t}
	default:
		return structRower{q, t}
	}
}

func (q dbQuery) Map() Curser {
	return mapRower{q}
}

func (q dbQuery) Cursor(args ...interface{}) Cursor {
	return q.Map().Cursor(args...)
}

///

type scanRower struct {
	q dbQuery
	t reflect.Type
}

func (r scanRower) Cursor(args ...interface{}) Cursor {
	return rows(r, r.q, args...)
}

func (r scanRower) scanRow(rows *sqlx.Rows) (T, error) {
	p := reflect.New(r.t).Interface()
	e := rows.Scan(p)
	return p, e
}

///

type structRower struct {
	q dbQuery
	t reflect.Type
}

func (r structRower) scanRow(rows *sqlx.Rows) (T, error) {
	p := reflect.New(r.t).Interface()
	e := rows.StructScan(p)
	return p, e
}

func (r structRower) Cursor(args ...interface{}) Cursor {
	return rows(r, r.q, args...)
}

///

type mapRower struct{ q dbQuery }

func (r mapRower) Cursor(args ...interface{}) Cursor {
	return rows(r, r.q, args...)
}

func (r mapRower) scanRow(rows *sqlx.Rows) (T, error) {
	m := make(map[string]interface{})
	e := rows.MapScan(m)
	return m, e
}

///

type rower interface {
	rows(Queryer, ...interface{}) Cursor
}

type rowScanner interface {
	scanRow(row *sqlx.Rows) (T, error)
}

type queryer interface {
	Exec(...interface{}) (sql.Result, error)
	Queryx(...interface{}) (*sqlx.Rows, error)
	QueryRowx(...interface{}) *sqlx.Row
}

///

func rows(s rowScanner, q dbQuery, args ...interface{}) Cursor {
	return do(rowCursor{MakeSignal(), s, q, args})
}

type rowCursor struct {
	Signal
	rowScanner
	dbQuery
	args []interface{}
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

func (q dbQuery) Exec(args ...interface{}) (sql.Result, error) {
	return q.db.Exec(q.q, args...)
}

func (q dbQuery) Queryx(args ...interface{}) (*sqlx.Rows, error) {
	return q.db.Queryx(q.q, args...)
}

func (q dbQuery) QueryRowx(args ...interface{}) *sqlx.Row {
	return q.db.QueryRowx(q.q, args...)
}
