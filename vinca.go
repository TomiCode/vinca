package main

import "log"
import "net/http"

var vincaConfig = VincaConfig{}

func main() {
    if vincaConfig.LoadConfig("config.json") != nil {
        return
    }
    log.Println("Starting vinca server..")

    if err := http.ListenAndServe(":3000", nil); err != nil {
        log.Fatal(err)
    }
}