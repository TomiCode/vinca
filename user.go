package main

import "log"
import "regexp"
import "golang.org/x/crypto/bcrypt"

type UserParam struct {
    Username string `json:"username"`
    Email string `json:"email"`
    Password string `json:"password,omitempty"`
}

type User struct {
    UserParam
    Id int  `json:"-"`
    Avatar string `json:"avatar"`
    LastUsed bool `json:"last_used"`
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

func (v *VincaDatabase) UserSave(usr *User) bool {
    hash, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
    if err != nil {
        log.Println("user save bcrypt err:", err)
        return false
    }

    res, err := v.db.Exec("insert into users(username, email, password) values(?,?,?)", usr.Username, usr.Email, hash)
    if err != nil {
        log.Println("user save db err:", err)
        return false
    }

    uid, err := res.LastInsertId()
    if err != nil {
        log.Println("user save id fetch err:", err)
    } else {
        usr.Id = int(uid)
    }

    return true
}

func (v *VincaDatabase) FetchUser(email string) *User {
    var usr = &User{}
    err := v.db.QueryRow("select * from users where email = ?", email).Scan(
        &usr.Id, &usr.Username, &usr.Email, &usr.hash, &usr.Avatar, &usr.LastUsed,
    )

    if err != nil {
        log.Println("unable to fetch user:", err)
        return nil
    }
    return usr
}

func (v *VincaDatabase) FetchUserFromSession(session *VincaSession) *User {
    var usr = &User{}
    err := v.db.QueryRow("select * from users where id = ?", session.userid).Scan(
        &usr.Id, &usr.Username, &usr.Email, &usr.hash, &usr.Avatar, &usr.LastUsed,
    )
    if err != nil {
        log.Println("no user for session:", err)
        return nil
    }
    return usr
}
