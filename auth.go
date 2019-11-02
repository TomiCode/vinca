package main

import "log"

type LoginResponse struct {
    Uuid string `json:"uuid"`
    User
}

func init() {
    vincaMux.NewRoute("/api/v1/auth/login").Handle(api_auth_login, "POST")
    vincaMux.NewRoute("/api/v1/auth/register").Handle(api_auth_register, "POST")
}

func api_auth_login(r *Request) interface{} {
    var params = UserParam{}
    if err := r.Decode(&params); err != nil {
        return err
    }

    usr := vincaDatabase.FetchUser(params.Email)
    if usr == nil {
        log.Println("unable to find requested user")
        return nil
    }

    if !usr.Authenticate(params.Password) {
        log.Println("nah, invalid password, try again")
        return nil
    }
    suid := vincaSessions.CreateSession(usr)

    return LoginResponse{Uuid: suid.String(), User: *usr}
}

func api_auth_register(r *Request) interface{} {
    var usr = User{}
    if err := r.Decode(&usr); err != nil {
        return err
    }
    log.Printf("%v\n", usr)

    if !usr.Valid() {
        log.Println("invalid user data received, try again")
        return nil
    }

    if !vincaDatabase.UserSave(&usr) {
        log.Println("failure while data save.")
        return nil
    }
    return usr
}