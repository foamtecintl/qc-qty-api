package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	createTable()
	http.HandleFunc("/save", saveData)
	http.HandleFunc("/check", checkStatus)
	http.HandleFunc("/unlock", unlock)
	fmt.Println("localhost:8090 is runing...")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

func getConnection() *sql.DB {
	dbConnect, err := sql.Open("sqlite3", "./database.sqlite")
	if err != nil {
		log.Fatalf("can not connect database : %v", err)
	}
	return dbConnect
}

func createTable() {
	db := getConnection()
	defer db.Close()
	sqlCreateTable := `
	CREATE TABLE IF NOT EXISTS BARCODE_COMPARE (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		createDate datetime DEFAULT NULL,
		part_master varchar(255) DEFAULT NULL,
		qty_master INTEGER DEFAULT NULL,
		status varchar(255) DEFAULT NULL,
		detail varchar(255) DEFAULT NULL,
		unlock_by varchar(255) DEFAULT NULL
	);`
	_, err := db.Exec(sqlCreateTable)
	if err != nil {
		panic(err)
	}
	sqlCreateTable = `
	CREATE TABLE IF NOT EXISTS USER_UNLOCK (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		createDate datetime DEFAULT NULL,
		name varchar(255) DEFAULT NULL,
		password varchar(255) DEFAULT NULL
	);`
	_, err = db.Exec(sqlCreateTable)
	if err != nil {
		panic(err)
	}
}

func bodyToJSON(r *http.Request) map[string]string {
	body := map[string]string{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		panic(err)
	}
	r.Body.Close()
	return body
}

func saveData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	db := getConnection()
	defer db.Close()
	if r.Method == http.MethodPost {
		body := bodyToJSON(r)
		t := time.Now()
		now := t.Format("2006-01-02 15:04:05")
		sqlQuery := `INSERT INTO BARCODE_COMPARE (
			createDate,
			part_master,
			qty_master,
			status,
			detail,
			unlock_by
		) VALUES (?,?,?,?,?,?)`
		_, err := db.Exec(
			sqlQuery,
			now,
			body["partMaster"],
			body["qtyMaster"],
			body["status"],
			body["detail"],
			body["unlockBy"],
		)
		if err != nil {
			panic(err)
		}
		mapData := make(map[string]string)
		mapData["status"] = "success"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mapData)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	return
}

func checkStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	db := getConnection()
	defer db.Close()
	if r.Method == http.MethodPost {
		sqlQuery := `SELECT ID, createDate, status, detail FROM BARCODE_COMPARE ORDER BY ID DESC LIMIT 1`
		row, err := db.Query(sqlQuery)
		if err != nil {
			panic(err)
		}
		var createDate, detail string
		status := "PASS"
		var id int
		for row.Next() {
			row.Scan(&id, &createDate, &status, &detail)
		}
		mapData := make(map[string]string)
		mapData["createDate"] = createDate
		mapData["status"] = status
		mapData["detail"] = detail
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mapData)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	return
}

func unlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	db := getConnection()
	defer db.Close()
	if r.Method == http.MethodPost {
		body := bodyToJSON(r)
		sqlQuery := `SELECT ID, createDate, name, password FROM USER_UNLOCK WHERE password = ?`
		row, err := db.Query(sqlQuery, body["password"])
		if err != nil {
			panic(err)
		}
		var createDate, name, password string
		var id int
		password = "no"
		for row.Next() {
			row.Scan(&id, &createDate, &name, &password)
		}
		mapData := make(map[string]string)
		if password == "no" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mapData)
			mapData["status"] = "fail"
			return
		}
		mapData["status"] = "success"
		mapData["name"] = name
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mapData)
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	return
}
