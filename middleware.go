package goup

import (
	"fmt"
	"github.com/startracex/goup/toolkit"
	"log"
	"runtime"
	"strings"
	"time"
)

// Logger record the request path, method, custom time
func Logger(flag ...int) HandlerFunc {
	d := 0
	for _, v := range flag {
		d = d | v
	}
	log.SetFlags(d)
	return func(req *HttpRequest, res *HttpResponse) {
		t := time.Now()
		req.Next(res)
		log.Printf("[%s] %s <%v>", req.Method, req.Path, time.Since(t))
	}
}

// Recovery error returns 500
func Recovery() HandlerFunc {
	return func(req *HttpRequest, res *HttpResponse) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				if req.Method == GET {
					res.ErrorStatusTextHTML(500)
				} else {
					res.Status(500)
				}
			}
		}()
		req.Next(res)
	}
}

// Cors adds multiple AllowOrigin or "*" and allows all other fields
func Cors(s ...string) HandlerFunc {
	c := toolkit.CorsAllowAll()
	if len(s) > 0 {
		c.AllowOrigin = append([]string{}, s...)
	}
	return SetCors(*c)
}

type CorsConfig = toolkit.Cors

// SetCors get the cors configuration from the parameter c
func SetCors(c CorsConfig) HandlerFunc {
	return func(req Request, res Response) {
		_ = c.WriteHeader(req.Header(), res.Header())
		req.Next(res)
	}
}

func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(4, pcs[:])
	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
