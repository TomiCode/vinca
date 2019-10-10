package main

import "log"
import "net/http"
import "encoding/json"


func init() {
    http.HandleFunc("/api/v1/auth/login", CorsFunc(api_auth_login))
    http.HandleFunc("/api/v1/auth/register", CorsFunc(api_auth_register))
}

func api_auth_login(w http.ResponseWriter, r *http.Request) {
    var params = UserParam{}
    if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
        log.Println("error:", err)
        return
    }
}

func api_auth_register(w http.ResponseWriter, r *http.Request) {
    var usr = User{}
    if err := json.NewDecoder(r.Body).Decode(&usr.UserParam); err != nil {
        log.Println("error:", err)
        return
    }
    log.Printf("%v\n", usr)

    if !usr.Valid() {
        log.Println("invalid user data received, try again")
        return
    }

    if !vincaDatabase.UserSave(&usr) {
        log.Println("failure while data save.")
        return
    }

    json.NewEncoder(w).Encode(true)
}