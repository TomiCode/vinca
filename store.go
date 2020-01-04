package main

import "log"
import "database/sql"

type Store struct {
    Id int `json:"id"`
    Created Datetime `json:"created"`
    LastUsed Datetime `json:"last_used"`
    Modified Datetime `json:"modified"`
    StoreParam
}

type StoreParam struct {
    Name string `json:"name"`
    Description string `json:"description"`
    Container int `json:"container"`
    Category int `json:"category"`
    Content []byte `json:"content"`
    Icon int `json:"icon"`
    Color int `json:"color"`
}

func (v *VincaDatabase) FetchStores(usr *User, sr StoresRequest) []Store {
    rows, err := v.db.Query("select id, container_id, category_id, created, last_used, modified, name, description, icon, color from stores where user_id = ? and category_id = ? order by name asc", usr.Id, sr.Category)
    if err != nil {
        log.Println("unable to fetch stores:", err)
        return nil
    }

    var stores []Store
    for rows.Next() {
        var st = Store{}
        err = rows.Scan(&st.Id, &st.Container, &st.Category,
                        &st.Created, &st.LastUsed, &st.Modified,
                        &st.Name, &st.Description, &st.Icon, &st.Color)
        if err != nil {
            log.Println("unable to scan single store:", err)
            continue
        }
        stores = append(stores, st)
    }

    log.Println("fetch stories for", usr.Username)
    return stores
}

func (v *VincaDatabase) FetchStoresWith(usr *User, params *CategoryRequest) []Store {
    var rows *sql.Rows = nil
    var err error = nil

    if params.Category == 0 {
        if params.Global == 1 {
            rows, err = v.db.Query("select id, container_id, category_id, created, last_used, modified, name, description, icon, color from stores where user_id = ? and category_id = 0", usr.Id)
        } else if params.Global == 2 {
            rows, err = v.db.Query("select id, container_id, category_id, created, last_used, modified, name, description, icon, color from stores where user_id = ? order by last_used limit 16", usr.Id)
        }
    } else {
        rows, err = v.db.Query("select id, container_id, category_id, created, last_used, modified, name, description, icon, color from stores where user_id = ? and category_id = ?", usr.Id, params.Category)
    }

    if err != nil || rows == nil {
        log.Println("unable to fetch stores:", err)
        return nil
    }

    var stores []Store
    for rows.Next() {
        var st = Store{}
        err = rows.Scan(&st.Id, &st.Container, &st.Category,
                        &st.Created, &st.LastUsed, &st.Modified,
                        &st.Name, &st.Description, &st.Icon, &st.Color)
        if err != nil {
            log.Println("unable to scan single store:", err)
            continue
        }

        stores = append(stores, st)
    }

    log.Println("fetch stories for", usr.Username)
    return stores
}

func (v *VincaDatabase) FetchStoreHistory(usr *User) []Store {
    if !usr.LastUsed {
        return []Store{ }
    }

    rows, err := v.db.Query("select id, container_id, category_id, created, last_used, modified, name, description, icon, color from stores where user_id = ? order by last_used desc limit 8", usr.Id)
    if err != nil {
        log.Println("unable to fetch stores:", err)
        return nil
    }

    var stores []Store
    for rows.Next() {
        var st = Store{}
        err = rows.Scan(&st.Id, &st.Container, &st.Category,
                        &st.Created, &st.LastUsed, &st.Modified,
                        &st.Name, &st.Description, &st.Icon, &st.Color)
        if err != nil {
            log.Println("unable to scan single store:", err)
            continue
        }
        stores = append(stores, st)
    }
    return stores
}

func (v *VincaDatabase) UpdateStoreUsage(usr *User, st *Store) {
    _, err := v.db.Exec("update stores set last_used = current_timestamp where id = ? and user_id = ?",
            st.Id, usr.Id)

    if err != nil {
        log.Println("error occurred while last_used update:", err)
    }
}

func (v *VincaDatabase) FetchStoreContent(usr *User, st *Store) error {
    row := v.db.QueryRow("select category_id, created, last_used, modified, name, description, icon, color, content from stores where id = ? and user_id = ?", st.Id, usr.Id)
    if err := row.Scan(&st.Category, &st.Created, &st.LastUsed,
        &st.Modified, &st.Name, &st.Description,
        &st.Icon, &st.Color, &st.Content); err != nil {

        log.Println("unable to fetch store content:", err)
        return err
    }

    if usr.LastUsed {
        go v.UpdateStoreUsage(usr, st)
    }

    log.Println("fetch store content for", usr.Username)
    return nil
}

func (v *VincaDatabase) SaveStore(usr *User, st *Store) error {
    res, err := v.db.Exec("insert into stores(user_id, container_id, category_id, name, description, icon, color, content) values(?,?,?,?,?,?,?,?)",
            usr.Id, st.Container, st.Category, st.Name, st.Description, st.Icon, st.Color, st.Content)

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

func (v *VincaDatabase) UpdateStore(usr *User, st *Store) error {
    res, err := v.db.Exec("update stores set category_id = ?, name = ?, description = ?, icon = ?, color = ?, content = ? where id = ? and user_id = ?",
            st.Category, st.Name, st.Description, st.Icon, st.Color, st.Content, st.Id, usr.Id)

    if err != nil {
        log.Println("unable to update store:", err)
        return err
    }

    rows, err := res.RowsAffected()
    if err != nil {
        log.Println("unable to fetch rows updated:", err)
        return err
    }

    if rows != 1 {
        log.Println("invalid rows updated!! Count:", rows)
    }
    return nil
}

func (v *VincaDatabase) DestroyStore(usr *User, st *Store) error {
    res, err := v.db.Exec("delete from stores where id = ? and user_id = ? limit 1", st.Id, usr.Id)
    if err != nil {
        log.Println("unable to remove store:", err)
        return err
    }

    rows, err := res.RowsAffected()
    if err != nil {
        log.Println("unable to fetch rows affected:", err)
        return err
    }
    if rows != 1 {
        log.Println("invalid number of deleted rows reported:", rows)
    }
    return nil
}