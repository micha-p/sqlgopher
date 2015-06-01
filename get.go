package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"regexp"
)

// retrieving column information might be combined better


func getConnection(cred Access, db string) *sql.DB {
	conn, err := sql.Open(cred.Dbms, dsn(cred.User, cred.Pass, cred.Host, cred.Port, db))
	checkY(err)
	return conn
}

func getRows(cred Access, db string, stmt string) *sql.Rows {
	conn := getConnection(cred, db)
	defer conn.Close()

	log.Println("[SQL] " + stmt)
	statement, err := conn.Prepare(stmt)
	checkY(err)
	rows, err := statement.Query()
	checkY(err)
	return rows
}


func getCols(cred Access, db string, t string) []string {

	conn := getConnection(cred, db)
	defer conn.Close()
	log.Println("[SQL] cols")
	// rows, err := conn.Query("select * from ? limit 1") // does not work??
	rows, err := conn.Query("select * from " + t + " limit 1")
	checkY(err)
	defer rows.Close()

	cols, err := rows.Columns()
	return cols
}

func getCount(cred Access, db string, t string) string {

	conn := getConnection(cred, db)
	defer conn.Close()
	log.Println("[SQL] count")
	// rows,err := conn.Query("select count(*) from ?", t) // does not work??
	rows,err := conn.Query("select count(*) from " + t)
	checkY(err)
	defer rows.Close()

	rows.Next()
	var field string
	rows.Scan(&field)
	return field
}

func getPrimary(cred Access, db string, t string) string {

	conn := getConnection(cred, db)
	defer conn.Close()
	// rows, err := conn.Query("show columns from ?", t) // does not work??
	rows, err := conn.Query("show columns from " + t)
	checkY(err)
	defer rows.Close()

	primary := ""
	for rows.Next() {
		var f, t, n, k, e string
		var d []byte // or use http://golang.org/pkg/database/sql/#NullString
		err := rows.Scan(&f, &t, &n, &k, &d, &e)
		checkY(err)
		if k == "PRI" {
			primary = f
		}
	}
	return primary
}


func getSingle(cred Access, db string, q string) string {

	rows := getRows(cred, db, q)
	defer rows.Close()
	var value interface{}
	var valuePtr interface{}
	valuePtr = &value

rowLoop:
	for rows.Next() {
		// just one row
		err := rows.Scan(valuePtr)
		checkY(err)
		break rowLoop
	}
	return dumpValue(value)
}


func getColumnInfo(cred Access, db string, t string) []CContext {

	conn := getConnection(cred, db)
	defer conn.Close()
	rows, err := conn.Query("show columns from " + t)
	checkY(err)
	defer rows.Close()

    m := []CContext{}
	for rows.Next() {
		var f, t, n, k, e string
		var d []byte // or use http://golang.org/pkg/database/sql/#NullString
		err := rows.Scan(&f, &t, &n, &k, &d, &e)
		checkY(err)

		iType, _ := regexp.MatchString("int", t)
		fType, _ := regexp.MatchString("float", t)
		rType, _ := regexp.MatchString("real", t)
		dType, _ := regexp.MatchString("double", t)
		lType, _ := regexp.MatchString("decimal", t)
		nType, _ := regexp.MatchString("numeric", t)
		cType, _ := regexp.MatchString("char", t)
		yType, _ := regexp.MatchString("binary", t)
		bType, _ := regexp.MatchString("blob", t)
		tType, _ := regexp.MatchString("text", t)

		if iType || fType || rType || dType || lType || nType {
			m=append(m,CContext{f,"numeric",""})
		} else if cType || yType || bType || tType {
			m=append(m,CContext{f,"","string"})
		} else {
			m=append(m,CContext{f,"",""})
		}
	}
	return m
}

