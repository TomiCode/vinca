package main

import "log"
import "net/http"
import "encoding/json"
import "github.com/google/uuid"

func init() {
    http.HandleFunc("/api/v1/home/container/create", CorsFunc(api_container_create))
    http.HandleFunc("/api/v1/home/container", CorsFunc(api_container_get))

    http.HandleFunc("/api/v1/home/categories", CorsFunc(api_categories))
    http.HandleFunc("/api/v1/home/category/create", CorsFunc(api_category_create))
    http.HandleFunc("/api/v1/home/category", CorsFunc(api_category_get))

    http.HandleFunc("/api/v1/home/stores", CorsFunc(api_stores))
    http.HandleFunc("/api/v1/home/store/content", CorsFunc(api_store_content))
    http.HandleFunc("/api/v1/home/store/create", CorsFunc(api_store_create))
}

type ContainerResponse struct {
    Valid bool `json:"valid"`
    Container
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

func api_container_get(w http.ResponseWriter, r *http.Request) {
    suid, err := uuid.Parse(r.Header.Get("Vinca-Authentication"))
    if err != nil {
        log.Println("unable to fetch session for user:", err)
        return
    }

    usr := vincaSessions.SessionUser(suid)
    if usr == nil {
        log.Println("invalid user session, try again")
        return
    }

    container := vincaDatabase.FetchContainer(usr)
    if container == nil {
        json.NewEncoder(w).Encode(ContainerResponse{Valid: false})
        return
    }
    json.NewEncoder(w).Encode(ContainerResponse{Valid: true, Container: *container})
}

func api_container_create(w http.ResponseWriter, r *http.Request) {
    suid, err := uuid.Parse(r.Header.Get("Vinca-Authentication"))
    if err != nil {
        log.Println("unable to fetch session for user:", err)
        return
    }

    usr := vincaSessions.SessionUser(suid)
    if usr == nil {
        log.Println("invalid session user")
        return
    }

    var req = ContainerRequest{}
    if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Println("unable to parse container request:", err)
        return
    }

    var container = &Container{
        Name: "Default",
        Certificate: req.Certificate,
        Encrypted: req.Encrypted,
    }

    if err = vincaDatabase.SaveContainer(container, usr); err != nil {
        log.Println("unable to save container:", err)
        return
    }
    json.NewEncoder(w).Encode(container)
}

func api_categories(w http.ResponseWriter, r *http.Request) {
    suid, err := uuid.Parse(r.Header.Get("Vinca-Authentication"))
    if err != nil {
        log.Println("unable to fetch session for user:", err)
        return
    }

    usr := vincaSessions.SessionUser(suid)
    if usr == nil {
        log.Println("invalid session user")
        return
    }

    json.NewEncoder(w).Encode(vincaDatabase.FetchCategories(usr))
}

func api_category_get(w http.ResponseWriter, r *http.Request) {

}

func api_category_create(w http.ResponseWriter, r *http.Request) {
    suid, err := uuid.Parse(r.Header.Get("Vinca-Authentication"))
    if err != nil {
        log.Println("unable to fetch session for user:", err)
        return
    }

    usr := vincaSessions.SessionUser(suid)
    if usr == nil {
        log.Println("invalid session user")
        return
    }

    var category = Category{}
    if err = json.NewDecoder(r.Body).Decode(&category.CategoryRequest); err != nil {
        log.Println("unable to decode params:", err)
        return
    }

    if err = vincaDatabase.SaveCategory(&category, usr); err != nil {
        log.Println("unable to save category to database:", err)
        return
    }
    json.NewEncoder(w).Encode(category)
}

func api_stores(w http.ResponseWriter, r *http.Request) {
    suid, err := uuid.Parse(r.Header.Get("Vinca-Authentication"))
    if err != nil {
        log.Println("unable to fetch session for user:", err)
        return
    }

    usr := vincaSessions.SessionUser(suid)
    if usr == nil {
        log.Println("invalid session user")
        return
    }

    json.NewEncoder(w).Encode(StoreResponse{
        Stores: vincaDatabase.FetchStores(usr),
    })
}

func api_store_content(w http.ResponseWriter, r *http.Request) {
    var param = StoreContentRequest{}
    if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
        log.Println("unable to fetch store_id:", err)
        return
    }

    suid, err := uuid.Parse(r.Header.Get("Vinca-Authentication"))
    if err != nil {
        log.Println("unable to fetch session for user:", err)
        return
    }

    usr := vincaSessions.SessionUser(suid)
    if usr == nil {
        log.Println("invalid session user")
        return
    }

    var localStore = Store{Id: param.StoreId}
    if err := vincaDatabase.FetchStoreContent(usr, &localStore); err != nil {
        log.Println("unable to fetch store:", err)
        return
    }
    json.NewEncoder(w).Encode(localStore)
}

func api_store_create(w http.ResponseWriter, r *http.Request) {
    var param = StoreParam{}
    if err := json.NewDecoder(r.Body).Decode(&param); err != nil {
        log.Println("unable to fetch store_id:", err)
        return
    }

    suid, err := uuid.Parse(r.Header.Get("Vinca-Authentication"))
    if err != nil {
        log.Println("unable to fetch session for user:", err)
        return
    }

    usr := vincaSessions.SessionUser(suid)
    if usr == nil {
        log.Println("invalid session user")
        return
    }

    var store = Store{StoreParam: param}
    if err = vincaDatabase.SaveStore(usr, &store); err != nil {
        log.Println("unable to save store to database.")
        return
    }

    json.NewEncoder(w).Encode(store)
}