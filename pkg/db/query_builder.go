package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type QueryBuilder struct {
	db        *sql.DB
	table     string
	columns   []string
	wheres    []string
	whereArgs []interface{}
	orderBy   string
	limit     int64
	offset    int64
}

func NewQueryBuilder(db *sql.DB) *QueryBuilder {
	return &QueryBuilder{db: db}
}

func (qb *QueryBuilder) Table(table string) *QueryBuilder {
	qb.table = table
	return qb
}

func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.columns = columns
	return qb
}

func (qb *QueryBuilder) Where(column string, operator string, value interface{}) *QueryBuilder {
	qb.wheres = append(qb.wheres, fmt.Sprintf("%s %s ?", column, operator))
	qb.whereArgs = append(qb.whereArgs, value)
	return qb
}

func (qb *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder {
	placeholders := strings.Repeat("?,", len(values))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	qb.wheres = append(qb.wheres, fmt.Sprintf("%s IN (%s)", column, placeholders))
	qb.whereArgs = append(qb.whereArgs, values...)
	return qb
}

func (qb *QueryBuilder) OrderBy(column string, direction string) *QueryBuilder {
	qb.orderBy = fmt.Sprintf("%s %s", column, direction)
	return qb
}

func (qb *QueryBuilder) Limit(limit int64) *QueryBuilder {
	qb.limit = limit
	return qb
}

func (qb *QueryBuilder) Offset(offset int64) *QueryBuilder {
	qb.offset = offset
	return qb
}

func (qb *QueryBuilder) buildSelect() (string, []interface{}) {
	var query strings.Builder
	var args []interface{}

	// SELECT columns
	if len(qb.columns) == 0 {
		query.WriteString("SELECT *")
	} else {
		query.WriteString("SELECT " + strings.Join(qb.columns, ", "))
	}

	// FROM table
	query.WriteString(" FROM " + qb.table)

	// WHERE conditions
	if len(qb.wheres) > 0 {
		query.WriteString(" WHERE " + strings.Join(qb.wheres, " AND "))
		args = append(args, qb.whereArgs...)
	}

	// ORDER BY
	if qb.orderBy != "" {
		query.WriteString(" ORDER BY " + qb.orderBy)
	}

	// LIMIT
	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}

	// OFFSET
	if qb.offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return query.String(), args
}

func (qb *QueryBuilder) Get() (*sql.Rows, error) {
	query, args := qb.buildSelect()
	return qb.db.Query(query, args...)
}

func (qb *QueryBuilder) First(dest ...interface{}) error {
	qb.Limit(1)
	query, args := qb.buildSelect()
	return qb.db.QueryRow(query, args...).Scan(dest...)
}

func (qb *QueryBuilder) Insert(data map[string]interface{}) (sql.Result, error) {
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		qb.table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return qb.db.Exec(query, values...)
}

func (qb *QueryBuilder) Update(data map[string]interface{}) (sql.Result, error) {
	var setParts []string
	var values []interface{}

	for col, val := range data {
		setParts = append(setParts, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}

	query := fmt.Sprintf("UPDATE %s SET %s", qb.table, strings.Join(setParts, ", "))

	if len(qb.wheres) > 0 {
		query += " WHERE " + strings.Join(qb.wheres, " AND ")
		values = append(values, qb.whereArgs...)
	}

	return qb.db.Exec(query, values...)
}

func (qb *QueryBuilder) Delete() (sql.Result, error) {
	query := fmt.Sprintf("DELETE FROM %s", qb.table)

	if len(qb.wheres) > 0 {
		query += " WHERE " + strings.Join(qb.wheres, " AND ")
	}

	return qb.db.Exec(query, qb.whereArgs...)
}
