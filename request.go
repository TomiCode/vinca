package main

import "net/http"

func CorsFunc(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        if r.Method != "OPTIONS" {
            handler(w, r)
            return
        }
        var headers = w.Header()
        headers.Add("Vary", "Origin")
        headers.Add("Vary", "Access-Control-Request-Method")
        headers.Add("Vary", "Access-Control-Request-Headers")
        headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, Aurora-Auth")
        headers.Add("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
    }
}