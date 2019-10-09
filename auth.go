package main

import "log"
import "net/http"
import "encoding/json"

type AuthLoginReq struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type AuthRegisterReq struct {
    Username string `json:"username"`
    Email string `json:"email"`
    Password string `json:"password"`
}

func init() {
    http.HandleFunc("/api/v1/auth/login", CorsFunc(api_auth_login))
    http.HandleFunc("/api/v1/auth/register", CorsFunc(api_auth_register))
}

func api_auth_login(w http.ResponseWriter, r *http.Request) {
    var request = AuthLoginReq{}
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        log.Println("error:", err)
        return
    }
}

func api_auth_register(w http.ResponseWriter, r *http.Request) {
    var request = AuthRegisterReq{}
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        log.Println("error:", err)
        return
    }
    log.Printf("%v\n", request)
}