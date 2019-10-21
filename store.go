package main

import "log"
import "time"

type Store struct {
    Id int `json:"id"`
    Created time.Time `json:"created"`
    LastUsed time.Time `json:"last_used"`
    Modified time.Time `json:"modified"`
    StoreParam
}

type StoreParam struct {
    Container int `json:"container"`
    Category int `json:"category"`
    Name string `json:"name"`
    Icon string `json:"icon"`
    Color int `json:"color"`
    Content []byte `json:"content"`
}

func (v *VincaDatabase) FetchStores(usr *User) []Store {
    rows, err := v.db.Query("select id, container_id, category_id, created, last_used, modified, name, icon, color from stores where user_id = ?", usr.Id)
    if err != nil {
        log.Println("unable to fetch stores:", err)
        return nil
    }

    var stores []Store
    for rows.Next() {
        var st = Store{}
        err = rows.Scan(&st.Id, &st.Container, &st.Category,
                        &st.Created, &st.LastUsed, &st.Modified,
                        &st.Name, &st.Icon, &st.Color)
        if err != nil {
            log.Println("unable to scan single store:", err)
            continue
        }

        stores = append(stores, st)
    }

    log.Println("fetch stories for", usr.Username)
    return stores
}

func (v *VincaDatabase) FetchStoreContent(usr *User, st *Store) error {
    err := v.db.QueryRow("select content from stores where id = ? and user_id = ?", st.Id, usr.Id).Scan(&st.Content)
    if err != nil {
        log.Println("unable to fetch store content:", err)
        return err
    }

    log.Println("fetch store content for", usr.Username)
    return nil
}

func (v *VincaDatabase) SaveStore(usr *User, st *Store) error {
    res, err := v.db.Exec("insert into stores(user_id, container_id, category_id, name, icon, color, content) values(?,?,?,?,?,?,?)",
            usr.Id, st.Container, st.Category, st.Name, st.Icon, st.Color, st.Content)

    if err != nil {
        log.Println("unable to insert store:", err)
        return err
    }

    sid, err := res.LastInsertId()
    if err != nil {
        log.Println("unable to fetch last insert id:", err)
        return err
    }

    st.Id = int(sid)
    return nil
}
