package main

import "log"

type CategoryDestroyRequest struct {
    Id int `json:"id"`
    Migrate int `json:"migrate"`
}

type CategoryParams struct {
    Name string `json:"name"`
    Description string `json:"description"`
    Icon int `json:"icon"`
}

type Category struct {
    Id int `json:"id"`
    CategoryParams
}

func (v *VincaDatabase) FetchCategories(usr *User) []*Category {
    rows, err := v.db.Query("select id, name, description, icon from categories where user_id = ?", usr.Id)
    if err != nil {
        log.Println("unable to fetch categories for user", usr.Username)
        return nil
    }

    var categories []*Category
    for rows.Next() {
        var category = &Category{}
        if err = rows.Scan(&category.Id, &category.Name, &category.Description, &category.Icon); err != nil {
            log.Println("category fetch err:", err)
            continue
        }
        categories = append(categories, category)
    }
    return categories
}

func (v *VincaDatabase) FetchCategory(ct *Category, usr *User) error {
    err := v.db.QueryRow("select name, description, icon from categories where id = ? and user_id = ?",
            ct.Id, usr.Id).Scan(&ct.Name, &ct.Description, &ct.Icon)
    if err != nil {
        log.Println("unable to fetch category from database:", err)
        return err
    }
    return nil
}

func (v *VincaDatabase) SaveCategory(ct *Category, usr *User) error {
    res, err := v.db.Exec("insert into categories(user_id, name, description, icon) values(?,?,?,?)",
                        usr.Id, ct.Name, ct.Description, ct.Icon)
    if err != nil {
        log.Println("unable to save category to db:", err)
        return err
    }

    cid, err := res.LastInsertId()
    if err != nil {
        log.Println("unable to fetch id from table.")
        return nil
    }

    ct.Id = int(cid)
    return nil
}

func (v *VincaDatabase) UpdateCategory(ct *Category, usr *User) error {
    res, err := v.db.Exec("update categories set name = ?, description = ?, icon = ? where id = ? and user_id = ?",
            ct.Name, ct.Description, ct.Icon, ct.Id, usr.Id)

    if err != nil {
        log.Println("unable to update category:", err)
        return err
    }

    rows, err := res.RowsAffected()
    if err != nil {
        log.Println("Unable to fetch affected rows from db:", err)
        return nil
    }

    if rows != 1 {
        log.Println("category update affected different row count:", rows)
    }
    return nil
}

func (v *VincaDatabase) MigrateCategory(ct, migrate *Category, usr *User) error {
    if migrate.Id != 0 {
        if err := v.FetchCategory(migrate, usr); err != nil {
            log.Println("unable to migrate to invalid category:", err)
            return err
        }
    }

    res, err := v.db.Exec("update stores set category_id = ? where category_id = ? and user_id = ?",
            migrate.Id, ct.Id, usr.Id)
    if err != nil {
        log.Println("unable to move stores into migration category:", err)
        return err
    }

    rows, err := res.RowsAffected()
    if err != nil {
        log.Println("error while counting affected rows for category migration:", err)
        return nil
    }
    log.Println("category migration for", usr.Id, "affected", rows)
    return nil
}

func (v *VincaDatabase) DestroyCategory(ct *Category, usr *User) error {
    res, err := v.db.Exec("delete from categories where id = ? and user_id = ?", ct.Id, usr.Id)
    if err != nil {
        log.Println("unable to remove category from database:", err)
        return err
    }

    rows, err := res.RowsAffected()
    if err != nil {
        log.Println("unable to fetch rows remove count:", err)
        return nil
    }

    if rows != 1 {
        log.Println("something bad happened, different affected row count:", rows)
    }
    return nil
}
