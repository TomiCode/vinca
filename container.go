package main

import "log"

type Container struct {
    Id int `json:"id"`
    Name string `json:"name"`
    Certificate []byte `json:"certificate"`
    Encrypted []byte `json:"encrypted"`
}

// Currently we support only a single container per user.
func (v *VincaDatabase) FetchContainer(usr *User) *Container {
    row := v.db.QueryRow("select id, name, public, encrypted from containers where user_id = ? limit 1", usr.Id)

    var container = &Container{}
    if err := row.Scan(&container.Id, &container.Name, &container.Certificate, &container.Encrypted); err != nil {
        log.Println("unable to fetch user container:", err)
        return nil
    }
    return container
}

func (v *VincaDatabase) SaveContainer(container *Container, usr *User) error {
    res, err := v.db.Exec("insert into containers(user_id, name, public, encrypted) values(?,?,?,?)",
            usr.Id, container.Name, container.Certificate, container.Encrypted)
    if err != nil {
        log.Println("unable to save container:", err)
        return err
    }

    if cid, err := res.LastInsertId(); err == nil {
        container.Id = int(cid)
    } else {
        log.Println("unable to fetch insert id:", err)
    }
    return nil
}
