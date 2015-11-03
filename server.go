package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/delaemon/go-gibson-serial-number/app"

	"golang.org/x/net/netutil"
)

var (
	AccessTime        time.Time = time.Now()
	AccessLogTemplate           = `{{.RemoteAddr}} {{.ContentType}} {{.Method}} {{.Path}} {{.Query}} {{.Body}} {{.UserAgent}}`
)

type AccessLogLine struct {
	RemoteAddr  string
	ContentType string
	Path        string
	Query       string
	Method      string
	Body        string
	UserAgent   string
}

func AccessLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		bufbody := new(bytes.Buffer)
		bufbody.ReadFrom(req.Body)
		body := bufbody.String()
		line := AccessLogLine{
			req.RemoteAddr,
			req.Header.Get("Content-Type"),
			req.URL.Path,
			req.URL.RawQuery,
			req.Method, body, req.UserAgent(),
		}
		tmpl, err := template.New("line").Parse(AccessLogTemplate)
		if err != nil {
			panic(err)
		}
		bufline := new(bytes.Buffer)
		err = tmpl.Execute(bufline, line)
		if err != nil {
			panic(err)
		}

		logFile := fmt.Sprintf("./log/access/%d%02d%02d.log", AccessTime.Year(), AccessTime.Month(), AccessTime.Day())
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		log.SetOutput(f)
		log.Printf(bufline.String())

		handler.ServeHTTP(w, req)
	})
}

func Server() {
	http.HandleFunc("/", app.Handler)
	defaultPort := 80
	defaultLimit := 50
	port := defaultPort
	limit := defaultLimit
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.IntVar(&port, "p", defaultPort, "listen port")
	f.IntVar(&port, "port", defaultPort, "listen port")
	f.IntVar(&limit, "l", defaultLimit, "server limit")
	f.IntVar(&limit, "limit", defaultLimit, "server limit")
	f.Parse(os.Args[1:])
	for 0 < f.NArg() {
		f.Parse(f.Args()[1:])
	}
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}

	limit_listener := netutil.LimitListener(listener, limit)
	defer limit_listener.Close()

	http_config := &http.Server{
		Handler: AccessLog(http.DefaultServeMux),
	}
	err = http_config.Serve(limit_listener)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	Server()
}
