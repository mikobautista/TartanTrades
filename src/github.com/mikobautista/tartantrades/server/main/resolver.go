package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"regexp"

	"github.com/mikobautista/tartantrades/logging"
	"github.com/mikobautista/tartantrades/rpc/resolver_rpc"
)

var (
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
	LOG  = logging.NewLogger(true)
)

type Page struct {
	Title string
	Body  []byte
}

type resolver struct {
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	// First initialize http server
	flag.Parse()
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	if *addr {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("final-port.txt", []byte(l.Addr().String()), 0644)
		if err != nil {
			log.Fatal(err)
		}
		s := &http.Server{}
		s.Serve(l)
		return
	}

	http.ListenAndServe(":8080", nil)

	// Initialize RPC server for trade servers
	r := &resolver{}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", 8081))
	err = rpc.RegisterName("StorageServer", resolver_rpc.Wrap(r))
	LOG.CheckForError(err, true)
	rpc.HandleHTTP()
	go http.Serve(ln, nil)

}

func (r *resolver) RegisterServer(*resolver_rpc.RegisterArgs, *resolver_rpc.RegisterReply) error {
	return errors.New("Not Implemented")
}

func (r *resolver) GetServers(*resolver_rpc.GetServersArgs, *resolver_rpc.GetServersReply) error {
	return errors.New("Not Implemented")
}

func (r *resolver) Get(*resolver_rpc.GetArgs, *resolver_rpc.GetReply) error {
	return errors.New("Not Implemented")
}

func (r *resolver) GetList(*resolver_rpc.GetArgs, *resolver_rpc.GetListReply) error {
	return errors.New("Not Implemented")
}

func (r *resolver) Put(*resolver_rpc.PutArgs, *resolver_rpc.PutReply) error {
	return errors.New("Not Implemented")
}

func (r *resolver) AppendToList(*resolver_rpc.PutArgs, *resolver_rpc.PutReply) error {
	return errors.New("Not Implemented")
}

func (r *resolver) RemoveFromList(*resolver_rpc.PutArgs, *resolver_rpc.PutReply) error {
	return errors.New("Not Implemented")
}
