package main

import "log"

type CategoryRequest struct {
    Title string `json:"title"`
    Description string `json:"description"`
    Icon string `json:"icon"`
}

type Category struct {
    Id int `json:"id"`
    CategoryRequest
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
        if err = rows.Scan(&category.Id, &category.Title, &category.Description, &category.Icon); err != nil {
            log.Println("category fetch err:", err)
            continue
        }
        categories = append(categories, category)
    }
    return categories
}

func (v *VincaDatabase) SaveCategory(ct *Category, usr *User) error {
    res, err := v.db.Exec("insert into categories(user_id, name, description, icon) values(?,?,?,?)",
                        usr.Id, ct.Title, ct.Description, ct.Icon)
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
