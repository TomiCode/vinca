package main

import "log"
import "database/sql"
import "regexp"
import "golang.org/x/crypto/bcrypt"

type UserParam struct {
    Username string `json:"username"`
    Email string `json:"email"`
    Password string `json:"password,omitempty"`
    LastUsed bool `json:"last_used"`
    DarkMode bool `json:"dark_mode"`
}

type User struct {
    UserParam
    Id int `json:"-"`
    Avatar string `json:"avatar"`
    hash []byte
}

var RgxUsernameCheck = regexp.MustCompile("^[A-Za-z]{1,16}$")
var RgxEmailCheck = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (p *UserParam) Valid() bool {
    if !RgxEmailCheck.MatchString(p.Email) || !RgxUsernameCheck.MatchString(p.Username)  {
        return false
    }
    return true
}

func (usr *User) Authenticate(password string) bool {
    err := bcrypt.CompareHashAndPassword(usr.hash, []byte(password))
    if err != nil {
        log.Println("auth err:", err)
        return false
    }
    return true
}

func (usr *User) SetPassword(password string) error {
    var err error

    usr.hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Println("user save bcrypt err:", err)
        return err
    }
    return nil
}

func (v *VincaDatabase) UserSave(usr *User) error {
    if err := usr.SetPassword(usr.Password); err != nil {
        return err
    }

    res, err := v.db.Exec("insert into users(username, email, password) values(?,?,?)",
            usr.Username, usr.Email, usr.hash)
    if err != nil {
        log.Println("user save db err:", err)
        return err
    }

    uid, err := res.LastInsertId()
    if err != nil {
        log.Println("user save id fetch err:", err)
    } else {
        usr.Id = int(uid)
    }
    return nil
}

func (v *VincaDatabase) FetchUser(email string) *User {
    var usr = &User{}
    err := v.db.QueryRow("select id, username, email, password, avatar, show_last_used, dark_mode from users where email = ?", email).Scan(
        &usr.Id, &usr.Username, &usr.Email, &usr.hash, &usr.Avatar, &usr.LastUsed, &usr.DarkMode,
    )

    if err != nil {
        log.Println("unable to fetch user:", err)
        return nil
    }
    return usr
}

func (v *VincaDatabase) FetchUserFromSession(session *VincaSession) *User {
    var usr = &User{}
    err := v.db.QueryRow("select id, username, email, password, avatar, show_last_used, dark_mode from users where id = ?", session.userid).Scan(
        &usr.Id, &usr.Username, &usr.Email, &usr.hash, &usr.Avatar, &usr.LastUsed, &usr.DarkMode,
    )
    if err != nil {
        log.Println("no user for session:", err)
        return nil
    }
    return usr
}

func (v *VincaDatabase) UpdateUser(usr *User, params UserParam) error {
    if usr.Email != params.Email {
        var uid int
        err := v.db.QueryRow("select id from users where email = ?", params.Email).Scan(&uid)

        if err != nil && err != sql.ErrNoRows {
            log.Println("error while email check:", err)
            return err
        } else if err == nil {
            return ErrUsedEmail
        }
    }

    if params.Password != "" {
        if err := usr.SetPassword(params.Password); err != nil {
            return err
        }
    }

    _, err := v.db.Exec("update users set email = ?, password = ?, show_last_used = ?, dark_mode = ? where id = ?",
            params.Email, usr.hash, params.LastUsed, params.DarkMode, usr.Id)
    if err != nil {
        log.Println("unable to update user properties:", err)
        return err
    }
    usr.DarkMode = params.DarkMode
    usr.LastUsed = params.LastUsed
    usr.Email = params.Email

    return nil
}
