package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	  dataDir     = "data/"
    templateDir = "tmpl/"
    templates   = template.Must(template.ParseFiles(templateDir+"edit.html", templateDir+"view.html"))
    validPath   = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

    fileNames []string
)

type Page struct {
    Title string
    Body  []byte
}

func fileName(title string) string {
	return dataDir+title+".txt"
}

func (p *Page) save() error {
    filename := fileName(p.Title)
    return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    filename := fileName(title)
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
    	http.Error(w, err.Error(), http.StatusInternalServerError)
    	return
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
    	http.Redirect(w, r, "/edit/"+title, http.StatusFound)
    	return
    }
    for _, name := range fileNames {
    	regx := regexp.MustCompile(fmt.Sprintf("(%s)", name))
    	p.Body = regx.ReplaceAll(p.Body, []byte("<a href=\"/view/"+name+"\">"+name+"</a>"))
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

func rootHandler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
			m := validPath.FindStringSubmatch(r.URL.Path)
			if m == nil {
        http.NotFound(w, r)
        return
      }
      fn(w, r, m[2])
  }
}

func fileNameWithoutExtension(fileName string) string {
	  return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func loadFileNames() error {
		files, err := ioutil.ReadDir("data")
		if err != nil {
	    return err
		}
		for _, f := range files {
		    fileNames = append(fileNames, fileNameWithoutExtension(f.Name()))
		}
		return nil

}

func main() {
	  loadFileNames()
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    http.HandleFunc("/", rootHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}