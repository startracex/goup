package goup

import (
	"bytes"
	"github.com/startracex/goup/toolkit"
	"html/template"
	"net/http"
	"strings"
	"sync"
)

type Engine struct {
	*RouterGroup
	router          *Router
	groups          []*RouterGroup
	Pool            sync.Pool
	Template        template.Template
	FuncMap         template.FuncMap
	noRouteHandler  []HandlerFunc
	noMethodHandler []HandlerFunc
}

// ServeHTTP for http.ListenAndServe
func (e *Engine) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix+"/") {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	q := NewRequest(req)
	q.Handlers = middlewares
	q.Engine = e
	p := NewResponse(res)
	p.Engine = e
	e.router.Handle(&q, &p)
}

// New create engine
func New() *Engine {
	engine := &Engine{router: NewRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	engine.Pool.New = func() any {
		return bytes.NewBuffer(make([]byte, 4096))
	}
	engine.NoRoute(func(request Request, response Response) {
		if request.Method == "GET" {
			response.ErrorStatusTextHTML(404)
			return
		}
		response.ErrorStatusText(404)
	})
	engine.NoMethod(func(request Request, response Response) {
		if request.Method == "GET" {
			response.ErrorStatusTextHTML(405)
			return
		}
		response.ErrorStatusText(405)
	})
	return engine
}

// Default use default middleware (Logger and Recovery)
func Default() *Engine {
	engine := New()
	engine.Use(Logger(1), Recovery())
	return engine
}

// SetPoolNew Replace the default NEW func
func (e *Engine) SetPoolNew(f func() any) {
	e.Pool.New = f
}

// LoadHTML load all files under paths
func (e *Engine) LoadHTML(path ...string) {
	var files []string
	for _, v := range path {
		files = append(files, toolkit.WalkFiles(v)...)
	}
	e.Template = *template.Must(template.New("").Funcs(e.FuncMap).ParseFiles(files...))
}

// LoadHTMLFiles load the path file
func (e *Engine) LoadHTMLFiles(path ...string) {
	e.Template = *template.Must(template.New("").Funcs(e.FuncMap).ParseFiles(path...))
}

// LoadFunc load func map for template
func (e *Engine) LoadFunc(funcMap template.FuncMap) {
	e.FuncMap = funcMap
}

// ListenAndServe start a server
func (e *Engine) ListenAndServe(addr string) error {
	if len(addr) > 0 && addr[0] != ':' {
		addr = ":" + addr
	}
	return http.ListenAndServe(addr, e)
}

// Run call ListenAndServe or ListenAndServeTLS if it has filePath slice
func (e *Engine) Run(addr string, filePath ...string) error {
	if len(filePath) > 1 {
		return e.ListenAndServeTLS(addr, filePath[0], filePath[1])
	}
	return e.ListenAndServe(addr)
}

// ListenAndServeTLS start a server with TLS
func (e *Engine) ListenAndServeTLS(addr, cert, key string) error {
	if len(addr) > 0 && addr[0] != ':' {
		addr = ":" + addr
	}
	return http.ListenAndServeTLS(addr, cert, key, e)
}
