package main

import "github.com/google/uuid"
import "net/http"
import "log"

const AuthSessionUser int = 0x1001

var ErrInvalidLogin = NewHandlerErr("user_login_invalid", http.StatusUnauthorized)
var ErrInvalidData = NewHandlerErr("user_data_invalid", http.StatusBadRequest)
var ErrInvalidSession = NewHandlerErr("user_session_invalid", http.StatusUnauthorized)

type LoginResponse struct {
    Uuid string `json:"uuid"`
    User
}

func init() {
    vincaMux.NewRoute("/api/v1/auth/login").Handle(api_auth_login, "POST")
    vincaMux.NewRoute("/api/v1/auth/register").Handle(api_auth_register, "POST")
    vincaMux.NewRoute("/api/v1/auth/reset").Handle(api_auth_reset, "POST")
    vincaMux.NewRoute("/api/v1/auth/session").Middleware(auth_middleware).Handle(api_auth_session, "GET")
}

func api_auth_login(r *Request) interface{} {
    var params = UserParam{}
    if err := r.Decode(&params); err != nil {
        return err
    }

    usr := vincaDatabase.FetchUser(params.Email)
    if usr == nil {
        log.Println("unable to find requested user")
        return ErrInvalidLogin
    }

    if !usr.Authenticate(params.Password) {
        log.Println("nah, invalid password, try again")
        return ErrInvalidLogin
    }
    suid := vincaSessions.CreateSession(usr)

    return LoginResponse{Uuid: suid.String(), User: *usr}
}

func api_auth_register(r *Request) interface{} {
    var usr = User{}
    if err := r.Decode(&usr.UserParam); err != nil {
        return err
    }

    if !usr.Valid() {
        log.Println("invalid user data received, try again")
        return ErrInvalidData
    }

    if err := vincaDatabase.UserSave(&usr); err != nil {
        return err
    }
    return usr
}

func api_auth_reset(r *Request) interface{} {
    var usr = User{}
    if err := r.Decode(&usr.UserParam); err != nil {
        return err
    }
    log.Println("Restore password:", usr)

    return usr
}

func api_auth_session(r *Request) interface{} {
    if usr, valid := r.Value(AuthSessionUser).(*User); valid {
        return usr
    }
    return ErrInvalidSession
}

func auth_middleware(r *Request) error {
    log.Println("auth_middleware")
    suid, err := uuid.Parse(r.Header.Get("Vinca-Authentication"))
    if err != nil {
        return err
    }

    if usr := vincaSessions.SessionUser(suid); usr != nil {
        r.WithValue(AuthSessionUser, usr)
        return nil
    }
    return ErrInvalidSession
}
