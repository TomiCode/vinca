package main

import "log"
import "net/http"

var sys_config = Config{}

func main() {
    if sys_config.LoadConfig("config.json") != nil {
        return
    }
    log.Println("Starting vinca server..")

    if err := http.ListenAndServe(":3000", nil); err != nil {
        log.Fatal(err)
    }
}