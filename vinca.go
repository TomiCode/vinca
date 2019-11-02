package main

import "log"
import "net/http"

var vincaConfig = VincaConfig{}

var vincaDatabase = VincaDatabase{}

var vincaSessions = make(SessionContainer)

var vincaMux = &VincaMux{Cors: true}

func main() {
    if vincaConfig.LoadConfig("config.json") != nil {
        return
    }

    if !vincaDatabase.Open() {
        log.Println("unable to open database connection")
        return
    }

    log.Println("Starting vinca server..")
    if err := http.ListenAndServe(":3000", vincaMux); err != nil {
        log.Fatal(err)
    }
}