package main

import "log"

func init() {
    var route *VincaRoute

    route = vincaMux.NewRoute("/api/v1/home/container")
    route.Middleware(auth_middleware)
    route.Handle(api_container_get, "GET")
    route.Handle(api_container_create, "POST")

    route = vincaMux.NewRoute("/api/v1/home/categories")
    route.Middleware(auth_middleware)
    route.Handle(api_categories, "GET")

    route = vincaMux.NewRoute("/api/v1/home/category")
    route.Middleware(auth_middleware)
    route.Handle(api_category_get, "GET")
    route.Handle(api_category_create, "POST")
    route.Handle(api_category_update, "PATCH")

    route = vincaMux.NewRoute("/api/v1/home/category/delete")
    route.Middleware(auth_middleware)
    route.Handle(api_category_remove, "POST")

    route = vincaMux.NewRoute("/api/v1/home/stores")
    route.Middleware(auth_middleware)
    route.Handle(api_stores, "POST")

    route = vincaMux.NewRoute("/api/v1/home/store/create")
    route.Middleware(auth_middleware)
    route.Handle(api_store_create, "POST")

    route = vincaMux.NewRoute("/api/v1/home/store")
    route.Middleware(auth_middleware)
    route.Handle(api_store_content, "POST")
    route.Handle(api_store_update, "PATCH")

    route = vincaMux.NewRoute("/api/v1/home/store/delete")
    route.Middleware(auth_middleware)
    route.Handle(api_store_remove, "POST")

    route = vincaMux.NewRoute("/api/v1/home")
    route.Middleware(auth_middleware)
    route.Handle(api_home, "GET")

    route = vincaMux.NewRoute("/api/v1/home/preferences")
    route.Middleware(auth_middleware)
    route.Handle(api_user_update, "POST")
}

type ContainerResponse struct {
    Container
    Categories []*Category `json:"categories"`
}

type ContainerRequest struct {
    Encrypted []byte `json:"encrypted"`
    Certificate []byte `json:"certificate"`
}

type StoreContentRequest struct {
    StoreId int `json:"store_id"`
}

type StoreResponse struct {
    Stores []Store `json:"stores"`
}

type CategoryResponse struct {
    Created *Category `json:"created"`
    Categories []*Category `json:"categories"`
}

type CategoryDestroyResponse struct {
    Removed Category `json:"removed"`
    Migrated Category `json:"migrated"`
}

type CategoryRequest struct {
    Category int `json:"category"`
    Global int `json:"global,omitempty"`
}

type StoresRequest struct {
    Category int `json:"category"`
}

type HomeResponse struct {
    Unassigned []Store `json:"unassigned"`
    History []Store `json:"history"`
}

type UserUpdateRequest struct {
    Confirmation string `json:"confirmation"`
    UserParam
}

func api_container_get(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    return ContainerResponse{
        Container: vincaDatabase.FetchContainer(usr),
        Categories: vincaDatabase.FetchCategories(usr),
    }
}

func api_container_create(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    var req = ContainerRequest{}
    if err := r.Decode(&req); err != nil {
        return err
    }

    var container = &Container{
        Name: "Default",
        Certificate: req.Certificate,
        Encrypted: req.Encrypted,
    }

    if err := vincaDatabase.SaveContainer(container, usr); err != nil {
        log.Println("unable to save container:", err)
        return nil
    }
    return ContainerResponse{
        Container: *container,
        Categories: vincaDatabase.FetchCategories(usr),
    }
}

func api_categories(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    return vincaDatabase.FetchCategories(usr)
}

func api_category_get(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    var param = CategoryRequest{}
    if err := r.Decode(&param); err != nil {
        return err
    }

    return vincaDatabase.FetchStoresWith(usr, &param)
}

func api_category_create(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    var category = Category{}
    if err := r.Decode(&category.CategoryParams); err != nil {
        return err
    }

    if err := vincaDatabase.SaveCategory(&category, usr); err != nil {
        log.Println("unable to save category to database:", err)
        return nil
    }

    return CategoryResponse{Created: &category,
        Categories: vincaDatabase.FetchCategories(usr),
    }
}

func api_category_update(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    var category = Category{}
    if err := r.Decode(&category); err != nil {
        return err
    }

    if err := vincaDatabase.UpdateCategory(&category, usr); err != nil {
        return err
    }
    return category
}

func api_category_remove(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    var req = CategoryDestroyRequest{}
    if err := r.Decode(&req); err != nil {
        return err
    }

    var category = Category{Id: req.Id}
    if err := vincaDatabase.FetchCategory(&category, usr); err != nil {
        return err
    }

    var migrate = Category{Id: req.Migrate}
    if err := vincaDatabase.MigrateCategory(&category, &migrate, usr); err != nil {
        return err
    }

    if err := vincaDatabase.DestroyCategory(&category, usr); err != nil {
        return err
    }

    return CategoryDestroyResponse{
        Removed: category,
        Migrated: migrate,
    }
}

func api_stores(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    var params = StoresRequest{}
    if err := r.Decode(&params); err != nil {
        return err
    }

    return StoreResponse{
        Stores: vincaDatabase.FetchStores(usr, params),
    }
}

func api_store_content(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    var param = StoreContentRequest{}
    if err := r.Decode(&param); err != nil {
        return err
    }

    var localStore = Store{Id: param.StoreId}
    if err := vincaDatabase.FetchStoreContent(usr, &localStore); err != nil {
        log.Println("unable to fetch store:", err)
        return nil
    }
    return localStore
}

func api_store_create(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    var store = Store{}
    if err := r.Decode(&store.StoreParam); err != nil {
        return err
    }

    if err := vincaDatabase.SaveStore(usr, &store); err != nil {
        log.Println("unable to save store to database.")
        return nil
    }
    return store
}

func api_store_update(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    var store = Store{}
    if err := r.Decode(&store); err != nil {
        return err
    }

    var dbStore = Store{Id: store.Id}
    if err := vincaDatabase.FetchStoreContent(usr, &dbStore); err != nil {
        log.Println("unable to fetch store state:", err)
        return nil
    }

    if store.Content == nil {
        store.Content = dbStore.Content
    }

    if err := vincaDatabase.UpdateStore(usr, &store); err != nil {
        log.Println("unable to update store:", err)
        return nil
    }
    return store
}

func api_store_remove(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    var store = Store{}
    if err := r.Decode(&store); err != nil {
        return err
    }

    if err := vincaDatabase.FetchStoreContent(usr, &store); err != nil {
        log.Println("unable to fetch removed store:", err)
        return err
    }

    if err := vincaDatabase.DestroyStore(usr, &store); err != nil {
        log.Println("unable to remove store:", err)
        return err
    }
    return store
}

func api_home(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    return HomeResponse{
        Unassigned: vincaDatabase.FetchStores(usr, StoresRequest{Category: 0}),
        History: vincaDatabase.FetchStoreHistory(usr),
    }
}

func api_user_update(r *Request) interface{} {
    usr, ok := r.Value(AuthSessionUser).(*User)
    if !ok {
        return nil
    }

    var params = UserUpdateRequest{}
    if err := r.Decode(&params); err != nil {
        return err
    }

    if !usr.Authenticate(params.Confirmation) {
        return ErrInvalidPassword
    }

    if err := vincaDatabase.UpdateUser(usr, params.UserParam); err != nil {
        return err
    }
    return usr
}