package main

import "fmt"
import "log"
import "time"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

type VincaDatabase struct {
    db *sql.DB
}

type Datetime time.Time

func (dt *Datetime) Scan(v interface{}) error {
    if v == nil {
        *dt = Datetime(time.Unix(0, 0))
        return nil
    }

    if arr, ok := v.([]byte); ok {
        t, err := time.Parse("2006-01-02 15:04:05", string(arr))
        if err == nil {
            *dt = Datetime(t)
            return nil
        }
        return err
    }
    return fmt.Errorf("failed to scan Datetime")
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