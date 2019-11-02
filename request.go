package main

import "encoding/json"
import "net/http"
import "strings"
import "sync"
import "log"

type RouteHandler func(*Request) interface{}

type MiddlewareHandler func(*Request) error

type VincaMux struct {
    Cors bool
    mu sync.RWMutex
    routes map[string]*VincaRoute
}

type VincaRoute struct {
    mu sync.Mutex
    methods []*RouteMethod
    middleware []MiddlewareHandler
}

type RouteMethod struct {
    method string
    handler RouteHandler
    middleware []MiddlewareHandler
}

type Request struct {
    *http.Request
    store []RequestStoreParam
}

type RequestStoreParam struct {
    key interface{}
    value interface{}
}

func NewRequest(r *http.Request) *Request {
    return &Request{Request: r}
}

func (r *Request) Decode(v interface{}) error {
    return json.NewDecoder(r.Body).Decode(v)
}

func (r *Request) Value(key interface{}) interface{} {
    for _, sp := range r.store {
        if sp.key == key {
            return sp.value
        }
    }
    return nil
}

func (r *Request) WithValue(key, value interface{}) {
    r.store = append(r.store, RequestStoreParam{key: key, value: value})
}


func (vm *VincaMux) match(path string) *VincaRoute {
    log.Println("match route", path)

    vm.mu.RLock()
    defer vm.mu.RUnlock()

    if r, ok := vm.routes[path]; ok {
        return r
    }

    for p, r := range vm.routes {
        if strings.HasPrefix(path, p) {
            return r
        }
    }
    return nil
}

func (vm *VincaMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    route := vm.match(r.URL.Path)
    if route == nil {
        http.Error(w, "route not defined", http.StatusNotFound)
        return
    }

    if vm.Cors {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        if r.Method == "OPTIONS" {
            header := w.Header()
            header.Add("Vary", "Origin")
            header.Add("Vary", "Access-Control-Request-Method")
            header.Add("Vary", "Access-Control-Request-Headers")
            header.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, Vinca-Authentication")
            header.Add("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
            return
        }
    }

    r_method := route.match(r.Method)
    if r_method == nil {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req = NewRequest(r)
    for _, mid := range route.middleware {
        if err := mid(req); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    for _, mid := range r_method.middleware {
        if err := mid(req); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    resp := r_method.handler(req)
    if err, valid := resp.(error); valid {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    json.NewEncoder(w).Encode(resp)
}

func (vm *VincaMux) NewRoute(path string) *VincaRoute {
    var route = &VincaRoute{}

    vm.mu.Lock()
    defer vm.mu.Unlock()

    if vm.routes == nil {
        vm.routes = make(map[string]*VincaRoute)
    }

    if _, ok := vm.routes[path]; ok {
        panic("try to override existing route path")
    }
    vm.routes[path] = route

    return route
}

func (vr *VincaRoute) match(method string) *RouteMethod {
    if len(vr.methods) < 1 {
        return nil
    }

    for _, mt := range vr.methods {
        if mt.method == method {
            return mt
        }
    }
    return nil
}

func (vr *VincaRoute) Handle(h RouteHandler, m string) *VincaRoute {
    vr.mu.Lock()
    defer vr.mu.Unlock()

    if len(vr.methods) > 0 {
        for _, mt := range vr.methods {
            if mt.method == m {
                panic("try to assing existing method handler")
            }
        }
    }
    vr.methods = append(vr.methods, &RouteMethod{handler: h, method: m})

    return vr
}

func (vr *VincaRoute) Middleware(middleware ...MiddlewareHandler) *VincaRoute {
    vr.mu.Lock()
    defer vr.mu.Unlock()

    if len(vr.methods) > 0 {
        mt := vr.methods[len(vr.methods) - 1]
        mt.middleware = append(mt.middleware, middleware...)
    } else {
        vr.middleware = append(vr.middleware, middleware...)
    }

    return vr
}

func CorsFunc(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        if r.Method != "OPTIONS" {
            handler(w, r)
            return
        }
        var headers = w.Header()
        headers.Add("Vary", "Origin")
        headers.Add("Vary", "Access-Control-Request-Method")
        headers.Add("Vary", "Access-Control-Request-Headers")
        headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, Vinca-Authentication")
        headers.Add("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
    }
}
