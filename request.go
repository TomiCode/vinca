package main

import "encoding/json"
import "net/http"
import "strings"
import "sync"
import "log"

const ErrSuccess = "success"
const ErrSystem = "error"

var ErrInvalidParams = NewHandlerErr("sys_invalid_params", http.StatusBadRequest)

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

type Response struct {
    Status string `json:"status"`
    Content interface{} `json:"content,omitempty"`
    statusCode int
}

type HandlerErr struct {
    err string
    status int
}

func NewHandlerErr(err string, status int) *HandlerErr {
    return &HandlerErr{err: err, status: status}
}

func (err *HandlerErr) Error() string {
    return err.err
}

func (err *HandlerErr) Response() *Response {
    return &Response{Status: err.err, statusCode: err.status}
}

func (resp *Response) Write(w http.ResponseWriter) {
    w.WriteHeader(resp.statusCode)
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func NewRequest(r *http.Request) *Request {
    return &Request{Request: r}
}

func (r *Request) Decode(v interface{}) error {
    if err := json.NewDecoder(r.Body).Decode(v); err != nil {
        log.Println("unable to decode:", err)
        return ErrInvalidParams
    }
    return nil
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
            header.Add("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
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
            if hlerr, valid := err.(*HandlerErr); valid {
                hlerr.Response().Write(w)
                return
            }
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(Response{Status: ErrSystem, Content: err.Error()})
            return
        }
    }

    for _, mid := range r_method.middleware {
        if err := mid(req); err != nil {
            if hlerr, valid := err.(*HandlerErr); valid {
                hlerr.Response().Write(w)
                return
            }
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(Response{Status: ErrSystem, Content: err.Error()})
            return
        }
    }

    resp := r_method.handler(req)
    if err, valid := resp.(*HandlerErr); valid {
        log.Println(r.URL.Path, err.Error())
        err.Response().Write(w)
        return
    }
    if err, valid := resp.(error); valid {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(Response{Status: ErrSystem, Content: err.Error()})
        return
    }
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(Response{Status: "success", Content: resp})
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
