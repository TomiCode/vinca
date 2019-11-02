package main

import "github.com/google/uuid"
import "fmt"
import "log"

const AuthSessionUser int = 0x1001

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

func auth_middleware(r *Request) error {
    log.Println("auth_middleware")

    suid, err := uuid.Parse(r.Header.Get("Vinca-Authentication"))
    if err != nil {
        return err
    }

    usr := vincaSessions.SessionUser(suid)
    if usr == nil {
        log.Println("invalid session user")
        return fmt.Errorf("invalid session")
    }

    r.WithValue(AuthSessionUser, usr)
    return nil
}
