package my

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"math"
	"regexp"
	"strings"
)

type DB struct {
	io.Closer
	db *sql.DB
}
func (DB) New(path string) DB {
	db, err := sql.Open("sqlite3", path)
	PanicIf(err)
	return DB{db: db}
}
func (db DB) Close() error {
	return db.db.Close()
}
func (db DB) SelectOne(query string, params []any, values ...any) bool {
	rows, err := db.db.Query(query, params...)
	PanicIf(err)
	defer func() { Must(rows.Close()) }()
	if rows.Next() {
		Must(rows.Scan(values...))
		if rows.Next() { panic("multiple rows found") }
		return true
	}
	return false
}
func (db DB) SelectMany(
	table string,
	columns []string,
	where map[string]any,
	orderBy []string,
) []map[string]any {
	//goland:noinspection SqlNoDataSourceInspection
	query := fmt.Sprintf(
		"SELECT %s FROM %s",
		strings.Join(columns, ", "),
		table,
	)
	whereStmt, whereValues := formatWhere(where)
	query += whereStmt
	if len(orderBy) > 0 { query += " ORDER BY " + strings.Join(orderBy, ", ") }
	rows, err := db.db.Query(query, whereValues...)
	PanicIf(err)
	defer func() { Must(rows.Close()) }()
	res := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, 0, len(columns))
		for range columns { values = append(values, new(any)) }
		Must(rows.Scan(values...))
		valuesCopy := make(map[string]any)
		for index, column := range columns { valuesCopy[column] = *(values[index].(*any)) }
		res = append(res, valuesCopy)
	}
	return res
}
func (db DB) Insert(table string, columnValues map[string]any) int64 {
	values := make([]any, 0, len(columnValues))
	insertColumns := make([]string, 0, len(columnValues))
	insertValues := make([]string, 0, len(columnValues))
	for column, value := range columnValues {
		values = append(values, value)
		insertColumns = append(insertColumns, column)
		insertValues = append(insertValues, "?") // facepalm
	}

	//goland:noinspection SqlNoDataSourceInspection
	result, err := db.db.Exec(
		fmt.Sprintf(
			"INSERT INTO %s(%s) VALUES (%s)",
			table,
			strings.Join(insertColumns, ", "),
			strings.Join(insertValues, ", "),
		),
		values...,
	)
	PanicIf(err)
	id, err := result.LastInsertId()
	PanicIf(err)

	return id
}
func (db DB) InsertMany(table string, columns []string, rows [][]any, progressMessage string) error {
	const MaxNrVars = 999
	batchSize := MaxNrVars / len(columns)

	var progress *ProgressBar
	if progressMessage != "" {
		progress = (*ProgressBar)(nil).New(progressMessage, int64(math.Ceil(float64(len(rows)) / float64(batchSize))))
	}

	for len(rows) > 0 {
		var batch [][]any
		if len(rows) > batchSize {
			batch = rows[0:batchSize]
			rows = rows[batchSize:]
		} else {
			batch = rows
			rows = [][]any{}
		}

		values := make([]any, 0, len(columns)*len(batch))
		insertValues := make([]string, 0, len(batch))
		for _, row := range batch {
			if len(row) != len(columns) { panic("incorrect nr columns") }
			insertValues2 := make([]string, 0, len(columns))
			for _, value := range row {
				values = append(values, value)
				insertValues2 = append(insertValues2, "?") // facepalm
			}
			insertValues = append(insertValues, "(" + strings.Join(insertValues2, ", ") + ")")
		}

		//goland:noinspection SqlNoDataSourceInspection
		_, err := db.db.Exec(
			fmt.Sprintf(
				"INSERT INTO %s(%s) VALUES %s",
				table,
				"`" + strings.Join(columns, "`, `") + "`",
				strings.Join(insertValues, ", "),
			),
			values...,
		)
		if err != nil { return err }

		if progressMessage != "" { progress.Add() }
	}
	return nil
}
func (db DB) Upsert(table string, columnValues map[string]any) int64 {
	whereClause := make([]string, 0, len(columnValues))
	values := make([]any, 0, len(columnValues))
	for column, value := range columnValues {
		whereClause = append(whereClause, column+" = ?")
		values = append(values, value)
	}

	var id int64
	if //goland:noinspection SqlNoDataSourceInspection
	db.SelectOne(
		fmt.Sprintf("SELECT id FROM %s WHERE %s", table, strings.Join(whereClause, " AND ")),
		values,
		&id,
	) {
		return id
	}
	return db.Insert(table, columnValues)
}
func (db DB) Update(table string, where map[string]any, values map[string]any) error {
	valuesList := make([]any, 0, len(where) + len(values))

	updateClause := make([]string, 0, len(values))
	for column, value := range values {
		updateClause = append(updateClause, column + " = ?")
		valuesList = append(valuesList, value)
	}

	whereClause := make([]string, 0, len(where))
	for column, value := range where {
		whereClause = append(whereClause, column + " = ?")
		valuesList = append(valuesList, value)
	}

	//goland:noinspection SqlNoDataSourceInspection
	_, err := db.db.Exec(
		fmt.Sprintf(
			"UPDATE %s SET %s WHERE %s",
			table,
			strings.Join(updateClause, ", "),
			strings.Join(whereClause, " AND "),
		),
		valuesList...
	)

	return err
}
func (db DB) Delete(table string, where map[string]any) {
	whereClause, values := formatWhere(where)
	//goland:noinspection SqlNoDataSourceInspection
	_, err := db.db.Exec(
		fmt.Sprintf("DELETE FROM %s %s", table, whereClause),
		values...
	)
	PanicIf(err)
}

func (db DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.db.Exec(query, args...)
}

func formatWhere(where map[string]any) (string, []any) {
	var whereStmt string
	var whereValues []any
	if len(where) > 0 {
		whereColumns := make([]string, 0, len(where))
		whereValues = make([]any, 0, len(where))
		nonAlphanumeric := regexp.MustCompile("\\W")
		questionMark := regexp.MustCompile("\\?")
		for column, value := range where {
			if nonAlphanumeric.MatchString(column) {
				whereColumns = append(whereColumns, column)
				if questionMark.MatchString(column) {
					whereValues = append(whereValues, value)
				}
			} else {
				whereColumns = append(whereColumns, column + " = ?")
				whereValues = append(whereValues, value)
			}
		}
		whereStmt = " WHERE " + strings.Join(whereColumns, " AND ")
	}
	return whereStmt, whereValues
}
