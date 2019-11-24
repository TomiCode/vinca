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

    route = vincaMux.NewRoute("/api/v1/home/stores")
    route.Middleware(auth_middleware)
    route.Handle(api_stores, "POST")

    route = vincaMux.NewRoute("/api/v1/home/store")
    route.Middleware(auth_middleware)
    route.Handle(api_store_content, "GET")
    route.Handle(api_store_create, "POST")
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

type CategoryRequest struct {
    Category int `json:"category"`
    Global int `json:"global,omitempty"`
}

type StoresRequest struct {
    Category int `json:"category"`
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
    if err := r.Decode(&category); err != nil {
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

    var param = StoreParam{}
    if err := r.Decode(&param); err != nil {
        return err
    }

    var store = Store{StoreParam: param}
    if err := vincaDatabase.SaveStore(usr, &store); err != nil {
        log.Println("unable to save store to database.")
        return nil
    }

    return store
}
