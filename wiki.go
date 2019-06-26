package main

import (
    "io/ioutil"
    "net/http"
    "log"
    "html/template"
    "regexp"
    "path/filepath"
)

var templates = map[string]*template.Template{}



var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
    Title string
    Body []byte
    Menu []string
}

func (p *Page) save() error {
    filename := "data/" + p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates[tmpl].ExecuteTemplate(w, "layout.html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func loadPage(title string) (*Page, error) {
    filename := "data/" + title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    files, err := filepath.Glob("data/*.txt")

    filenameRegexp := regexp.MustCompile("^data/([^.]+).txt")
    filesNoPath := []string{}
    for _, file := range files {
        match := filenameRegexp.FindStringSubmatch(file)
        filesNoPath = append(filesNoPath, match[1])
    }

    if err != nil {
        log.Fatal(err)
    }
    return &Page{Title: title, Body: body, Menu: filesNoPath}, nil
}

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

func rootHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/" + title, http.StatusFound)
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
    http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func registerTemplates() {
    templates["view"] = template.Must(template.ParseFiles("templates/view.html", "templates/layout.html"))
    templates["edit"] = template.Must(template.ParseFiles("templates/edit.html", "templates/layout.html"))
}

func main() {
    // for the root do not use makeHandler as we are not interested in
    // validating the title. It is just a redirect
    registerTemplates()
    http.HandleFunc("/", rootHandler)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))

    log.Fatal(http.ListenAndServe(":8080", nil))
}
