package main

import "log"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

type VincaDatabase struct {
    db *sql.DB
}

func (vb *VincaDatabase) Open() bool {
    var err error
    vb.db, err = sql.Open("mysql", vincaConfig.Database)

    if err != nil {
        log.Println("Error occurred while database open:", err)
        return false
    }
    return true
}