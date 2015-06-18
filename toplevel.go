package main

import (
	"database/sql"
	"net/http"
	"net/url"
)

// rows are not adressable:
// dumpRows  -> SELECTFORM, INSERTFORM, INFO
// dumpRange -> SELECTFORM, INSERTFORM, INFO
// dumpField -> SELECTFORM, INSERTFORM, INFO

// rows are selected by where-clause
// dumpWhere 		-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO

// rows are selected by key or group
// dumpKeyValue 	-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO
// dumpGroup	 	-> SELECTFORM, INSERTFORM, UPDATEFORM, DELETE, INFO




func dumpIt(w http.ResponseWriter, r *http.Request, conn *sql.DB,
	host string, db string, t string, o string, d string, n string, g string, k string, v string) {

	if db == "" {
		dumpHome(w, conn, host)
		return
	} else if t == "" {
		dumpTables(w, conn, host, db, t, o, d, g, v)
	} else if k != "" && v != "" && k == getPrimary(conn, t) {
		dumpKeyValue(w, db, t, k, v, conn, host, sqlStar(t)+sqlWhere(k, "=", v))
	} else {
		dumpSelection(w, r, conn, host, db, t, o, d, n, g, k, v)
	}
}

// Shows selection of databases at top level
// TODO: Chnage to formUSE and actionUSE
func dumpHome(w http.ResponseWriter, conn *sql.DB, host string) {

	q := url.Values{}
    // "SELECT TABLE_NAME AS `Table`, ENGINE AS `Engine`, TABLE_ROWS AS `Rows`,TABLE_COLLATION AS `Collation`,CREATE_TIME AS `Create`, TABLE_COMMENT AS `Comment`
	stmt := string2sql("SHOW DATABASES")
	rows, err, _ := getRows(conn, stmt)
	checkY(err)
	defer rows.Close()

	records := [][]Entry{}
	head := []Entry{{"#", "", ""}, {"Database", "", ""}}
	var n int64
	for rows.Next() {
		n = n + 1
		var field string
		rows.Scan(&field)
		if EXPERTFLAG || INFOFLAG || field != "information_schema" {
			q.Set("db", field)
			link := q.Encode()
			row := []Entry{escape(Int64toa(n), link), escape(field, link)}
			records = append(records, row)
		}
	}
	// message suppressed, as it is not really useful and database should be chosen at login or bookmarked
	tableOutSimple(w, conn, host, "", "", head, records, []Entry{})
}

//  Dump all tables of a database
func dumpTables(w http.ResponseWriter, conn *sql.DB, host string, db string, t string, o string, d string, g string, v string) {

	q := url.Values{}
	q.Add("db", db)
	query := string2sql("SELECT TABLE_NAME AS `Table`, TABLE_ROWS AS `Rows`, TABLE_COMMENT AS `Comment`")

	query = query + " FROM information_schema.TABLES"
	query = query + sqlWhere("TABLE_SCHEMA","=",db) + sqlHaving(g, "=", v) + sqlOrder(o,d)
	rows, err, sec := getRows(conn, query)
	checkY(err)
	defer rows.Close()

	columns, err := rows.Columns()
	checkY(err)
	home := url.Values{}
	home.Add("db", db)
	home.Add("o", o)
	home.Add("d", d)
	head := createHead(db, "", o, d, "", "", columns, home)

	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	for i, _ := range columns {
		valuePtrs[i] = &values[i]
	}
	records := [][]Entry{}
	var rownum int64
	for rows.Next() {
		rownum = rownum + 1
		row := []Entry{}
		err = rows.Scan(valuePtrs...)
		checkY(err)
		g := url.Values{}
		g.Add("db", db)
		g.Add("t", getNullString(values[0]).String)
		row = append(row, escape(Int64toa(rownum), g.Encode()))
		for i, c := range columns {
			nv := getNullString(values[i])
			if c == "Rows" && (db == "INFORMATION_SCHEMA" || db =="information_schema") && (INFOFLAG || EXPERTFLAG) {
				nv = sql.NullString{Valid: true, String: getCount(conn,row[1].Text)}
			}
			if c == "Table" || c == "Comment" {
				v := nv.String
				g := url.Values{}
				g.Add("db", db)
				g.Add("t", v)
				row = append(row, escape(v, g.Encode()))
			} else {
				row = append(row, makeEntry(nv, db, "", c, ""))
			}
		}
		records = append(records, row)
	}

	// Shortened statement
	query = "SHOW TABLES" + sqlHaving(g, "=", v) + sqlOrder(o,d)
	var msg Message
	if QUIETFLAG {
		msg = Message{}
	} else {
		msg = Message{Msg:sql2string(query),Rows:rownum,Affected:-1,Seconds:sec }
	}
	tableOutRows(w, conn, host, db, "", "", "", "", "", "", Entry{}, Entry{}, head, records, []Entry{}, []Message{msg}, "", url.Values{})
}
