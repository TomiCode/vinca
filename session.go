package main

import "log"
import "github.com/google/uuid"

type VincaSession struct {
    userid int
}

type SessionContainer map[uuid.UUID]*VincaSession

func (sc SessionContainer) CreateSession(usr *User) uuid.UUID {
    log.Println("create session for user")
    suid, err := uuid.NewRandom()
    if err != nil {
        log.Println("unable to create session uuid:", err)
        return uuid.Nil
    }

    sc[suid] = &VincaSession{
        userid: usr.Id,
    }
    return suid
}

func (sc SessionContainer) SessionUser(suid uuid.UUID) *User {
    session, valid := sc[suid]
    if !valid {
        log.Println("unable to find session:", suid)
        return nil
    }

    usr := vincaDatabase.FetchUserFromSession(session)
    if usr == nil {
        log.Println("invalid user for session:", suid)
        return nil
    }
    return usr
}
