package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
	//"fmt"
)

type FContext struct {
	CSS      string
	Action   string
	Selector string
	Button   string
	Database string
	Table    string
	Order    string
	Desc     string
	Back     string
	Columns  []CContext
	Hidden   []CContext
	Trail    []Entry
	Level    string
}

func actionRouter(w http.ResponseWriter, r *http.Request, conn *sql.DB, host string, db string) {

	t, o, d, n, g, k, v := readRequest(r)

	q := r.URL.Query()
	action := q.Get("action")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	colinfo := getColumnInfo(conn, t)
	wclauses, sclauses := collectClauses(r, colinfo)
	// TODO change *FORM to form="*"
	if action == "INFO" && o =="" {
		stmt := str2sql("SHOW COLUMNS FROM ") + sqlProtectIdentifier(t)
		showColumns(w, conn, t, stmt)
	} else if action == "INFO" && o != "" {
		showStat(w, conn, t, o)
	} else if action == "GOTO" && n != "" {
		dumpRouter(w, r, conn, t, o, d, n, g, k, v)
	} else if action == "SELECT" {
		dumpRouter(w, r, conn, t, o, d, n, g, k, v)
	} else if action == "BACK" {
		dumpRouter(w, r, conn, "", "", "", "", "", "", "")
	} else if action == "SELECTFORM" {
		actionSELECTFORM(w, r, conn, t, o, d)
	} else if action == "INSERTFORM" && !READONLY {
		actionINSERTFORM(w, r, conn, t, o, d)
	} else if action == "DELETEFORM" && !READONLY {
		actionDELETEFORM(w, r, conn, t, o, d)

		// TODO check Update insert and delete
	} else if action == "UPDATEFORM" && !READONLY && k != "" && v != "" {
		actionKV_UPDATEFORM(w, r, conn, t, k, v)
	} else if action == "UPDATEFORM" && !READONLY && g != "" && v != "" {
		actionGV_UPDATEFORM(w, r, conn, t, g, v)
	} else if action == "UPDATEFORM" && !READONLY {
		actionUPDATEFORM(w, r, conn, t, o, d)

	} else if action == "INSERT" && !READONLY && len(sclauses) > 0 {
		stmt := sqlInsert(t) + sqlSetClauses(sclauses)
		actionEXEC(w, conn, t, o, d, n, g, v, wclauses, stmt)

	} else if action == "UPDATE" && !READONLY && k != "" && v != "" && len(sclauses) > 0 {
		hclause := []Clause{Clause{Column: k, Operator: "s?"}}
		stmt := sqlUpdate(t) + sqlSetClauses(sclauses) + sqlWhereClauses(append(wclauses, hclause))
		actionEXEC1(w, conn, t, o, d, n, g, v, wclauses, stmt, v)
	} else if action == "UPDATE" && !READONLY && g != "" && v != "" && len(sclauses) > 0 {
		hclause := []Clause{Clause{Column: g, Operator: "s?"}}
		stmt := sqlUpdate(t) + sqlSetClauses(sclauses) + sqlWhereClauses(append(wclauses, hclause))
		actionEXEC1(w, conn, t, o, d, n, g, v, wclauses, stmt, v)
	} else if action == "UPDATE" && !READONLY && len(sclauses) > 0 {
		stmt := sqlUpdate(t) + sqlSetClauses(sclauses) + sqlWhereClauses(wclauses)
		actionEXEC(w, conn, t, o, d, n, g, v, wclauses, stmt)

	} else if action == "DELETE" && !READONLY && g != "" && v != "" {
		hclause := []Clause{Clause{Column: g, Operator: "s?"}}
		stmt := sqlDelete(t) + sqlWhereClauses(append(wclauses, hclause))
		actionEXEC1(w, conn, t, o, d, n, g, v, [][]Clause{}, stmt, v)
	} else if action == "DELETE" && !READONLY && k != "" && v != "" {
		hclause := []Clause{Clause{Column: k, Operator: "s?"}}
		stmt := sqlDelete(t) + sqlWhereClauses(append(wclauses, hclause))
		actionEXEC1(w, conn, t, o, d, n, g, v, [][]Clause{}, stmt, v)
	} else if action == "DELETE" && !READONLY {
		stmt := sqlDelete(t) + sqlWhereClauses(wclauses)
		actionEXEC(w, conn, t, o, d, n, g, v, [][]Clause{}, stmt)

	} else {
		shipMessage(w, conn, "Action unknown or insufficient parameters: "+action)
	}
}

// INSERTFORM and SELECTFORM provide columns without values, EDIT/UPDATE provide a filled vmap
// TODO: use DEFAULT and AUTOINCREMENT

func shipForm(w http.ResponseWriter, r *http.Request, conn *sql.DB,
	t string, o string, d string,
	action string, button string, selector string, showncols []CContext, hiddencols []CContext, whereStack [][]Clause) {

	q := r.URL.Query()
	q.Del("action")
	linkback := q.Encode()

	host,db := getHostDB(getDSN(conn))
	c := FContext{
		CSS:      CSS_FILE,
		Action:   action,
		Selector: selector,
		Button:   button,
		Database: db,
		Table:    t,
		Order:    o,
		Desc:     d,
		Back:     linkback,
		Columns:  showncols,
		Hidden:   hiddencols,
		Trail:    makeTrail(host, db, t, "", "", whereStack),
		Level:    strconv.Itoa(len(whereStack) + 1),
	}

	if DEBUGFLAG {
		initTemplate()
	}
	err := templateForm.Execute(w, c)
	checkY(err)
}

/* The next four functions generate forms for doing SELECT, DELETE, INSERT, UPDATE */

func actionSELECTFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, t string, o string, d string) {
	colinfo := getColumnInfo(conn, t)
	whereStack := WhereQuery2Stack(r.URL.Query(), colinfo)
	hiddencols := WhereStack2Hidden(whereStack)
	shipForm(w, r, conn, t, o, d, "SELECT", "Select", "true", colinfo, hiddencols, whereStack)
}

func actionDELETEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, t string, o string, d string) {
	colinfo := getColumnInfo(conn, t)
	whereStack := WhereQuery2Stack(r.URL.Query(), colinfo)
	shipForm(w, r, conn, t, o, d, "DELETE", "Delete", "true", colinfo, []CContext{}, whereStack)
}

func actionINSERTFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, t string, o string, d string) {
	colinfo := getColumnInfo(conn, t)
	whereStack := WhereQuery2Stack(r.URL.Query(), colinfo)
	hiddencols := WhereStack2Hidden(whereStack)
	shipForm(w, r, conn, t, o, d, "INSERT", "Insert", "", colinfo, hiddencols, whereStack)
}

// TODO combine next 3 to 1 function: always promote gk,v, always fill if count = 1
func actionUPDATEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, t string, o string, d string) {

	colinfo := getColumnInfo(conn, t)
	wclauses, _ := collectClauses(r, colinfo)
	whereStack := WhereQuery2Stack(r.URL.Query(), colinfo)
	hiddencols := WhereStack2Hidden(whereStack)

	count, _ := getSingleValue(conn, sqlCount(t)+sqlWhereClauses(wclauses))
	if count == "1" {
		rows, err, _ := getRows(conn, sqlStar(t)+sqlWhereClauses(wclauses))
		checkY(err)
		defer rows.Close()
		shipForm(w, r, conn, t, o, d, "UPDATE", "Update", "", getColumnInfoFilled(conn, t, "", rows), hiddencols, whereStack)
	} else {
		shipForm(w, r, conn, t, o, d, "UPDATE", "Update", "", colinfo, hiddencols, whereStack)
	}
}
func actionKV_UPDATEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, t string, k string, v string) {
	colinfo := getColumnInfo(conn, t)
	whereStack := WhereQuery2Stack(r.URL.Query(), colinfo)
	col_g := findColumn(colinfo,k)
	isNumeric := col_g.IsNumeric
	var numeric bool
	if isNumeric != "" {
		numeric = true
	} else {
		numeric = false
	}
	whereStack = append(whereStack, []Clause{Clause{k, "=", v, numeric}})
	hiddencols := []CContext{
		CContext{"", "k", "", "", "", "", "valid", k, ""},
		CContext{"", "v", "", "", "", "", "valid", v, ""}}
	stmt := sqlStar(t) + sqlWhere1(k, "=")
	preparedStmt, _, err := sqlPrepare(conn, stmt)
	defer preparedStmt.Close()
	checkErrorPage(w, conn, t, stmt, err)
	rows, _, err := sqlQuery1(preparedStmt, v)
	checkY(err)
	defer rows.Close()
	primary := getPrimary(conn, t)
	shipForm(w, r, conn, t, "", "", "UPDATE", "Update", "", getColumnInfoFilled(conn, t, primary, rows), hiddencols, whereStack)
}
func actionGV_UPDATEFORM(w http.ResponseWriter, r *http.Request, conn *sql.DB, t string, g string, v string) {
	colinfo := getColumnInfo(conn, t)
	whereStack := WhereQuery2Stack(r.URL.Query(), colinfo)
	col_g := findColumn(colinfo,g)
	isNumeric := col_g.IsNumeric
	var numeric bool
	if isNumeric != "" {
		numeric = true
	} else {
		numeric = false
	}
	whereStack = append(whereStack, []Clause{Clause{g, "=", v, numeric}})
	hiddencols := []CContext{
		CContext{"", "g", "", "", "", "", "valid", g, ""},
		CContext{"", "v", "", "", "", "", "valid", v, ""}}
	stmt := sqlStar(t) + sqlWhere1(g, "=")
	preparedStmt, _, err := sqlPrepare(conn, stmt)
	checkErrorPage(w, conn, t, stmt, err)
	defer preparedStmt.Close()
	rows, _, err := sqlQuery1(preparedStmt, v)
	checkY(err)
	defer rows.Close()
	primary := getPrimary(conn, t)
	shipForm(w, r, conn, t, "", "", "UPDATE", "Update", "", getColumnInfoFilled(conn, t, primary, rows), hiddencols, whereStack)
}

// Excutes a statement on a selection by where-clauses
// Used, when rows are not adressable by a primary key or in table having a group
func actionEXEC(w http.ResponseWriter, conn *sql.DB, t string, o string, d string, n string, g string, v string,
	whereStack [][]Clause,stmt sqlstring) {

	messageStack := []Message{}
	preparedStmt, _, err := sqlPrepare(conn, stmt)
	checkErrorPage(w, conn, t, stmt, err)
	defer preparedStmt.Close()

	result, sec, err := sqlExec(preparedStmt)
	checkErrorPage(w, conn, t, stmt, err)
	affected, err := result.RowsAffected()
	checkErrorPage(w, conn, t, stmt, err)

	messageStack = append(messageStack, Message{sql2str(stmt), -1, affected, sec})
	nextstmt := sqlStar(t) + sqlWhereClauses(whereStack) + sqlOrder(o, d)
	dumpSelection(w, conn, t, o, d, n, g, v, nextstmt, whereStack, messageStack)
}

/* Executes prepared statements about modifications in tables with primary key or having a group
 * Uses one argument as value for where clause
 * However, prepared statements only work with values, not in identifier position */
func actionEXEC1(w http.ResponseWriter, conn *sql.DB, t string, o string, d string, n string, g string, v string,
	whereStack [][]Clause, stmt sqlstring, arg string) {

	messageStack := []Message{}
	preparedStmt, sec, err := sqlPrepare(conn, stmt)
	checkErrorPage(w, conn, t, stmt, err)
	defer preparedStmt.Close()
	messageStack = append(messageStack, Message{"PREPARE stmt FROM '" + sql2str(stmt) + "'", -1, 0, sec})

	result, sec, err := sqlExec1(preparedStmt, arg)
	checkErrorPage(w, conn, t, stmt, err)
	affected, err := result.RowsAffected()
	checkErrorPage(w, conn, t, stmt, err)

	messageStack = append(messageStack, Message{"EXECUTE stmt USING \"" + arg + "\"", -1, affected, sec})
	nextstmt := sqlStar(t) + sqlWhereClauses(whereStack) + sqlOrder(o, d)
	dumpSelection(w, conn, t, o, d, n, g, v, nextstmt, whereStack, messageStack)
}
