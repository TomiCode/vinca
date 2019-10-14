package main

import "log"

type Category struct {
    Id int `json:"id"`
    Title string `json:"title"`
    Description string `json:"description"`
    Icon string `json:"icon"`
}

func (v *VincaDatabase) FetchCategories(usr *User) []*Category {
    rows, err := v.db.Query("select id, title, description, icon from categories where user_id = ?", usr.Id)
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
