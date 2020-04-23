package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/markbates/pkger"
)

// template related methods for web struct.

func (web *web) initTemplates() {
	if web.templates != nil && !web.devMode {
		return
	}

	if web.templates == nil {
		web.templates = make(map[string]*template.Template)
	}

	// Force the include.
	pkger.Include("/templates")

	templatePaths, err := web.glob("/templates", "*.html")
	if err != nil {
		log.Fatal("Error initializing HTML Templates", err)
	}
	// templateHelperPath := path.Join("templates", "helper.html")

	funcMap := template.FuncMap{
		"FormatUnixDate": func(date int) string { return time.Unix(int64(date), 0).UTC().Format(time.RFC3339) },
		"CleanupXML":     func(xml string) string { return strings.ReplaceAll(xml, "&#xA;", "\n") },
	}
	_ = uint64(34)

	log.Printf("Loading %d templates from %v", len(templatePaths), "/templates")

	for _, filePath := range templatePaths {
		name := strings.TrimSuffix(path.Base(filePath), ".html")
		t := template.New(name).Funcs(funcMap)
		// TODO: Generalize the abstraction of Pkger?
		if web.devMode {
			web.templates[name] = template.Must(t.ParseFiles("./" + filePath /*, "./"+templateHelperPath*/))
		} else {
			web.templates[name] = template.Must(web.parseFiles(t, filePath /*, templateHelperPath*/))
		}
	}
}

func (web *web) renderTemplate(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {

	tmpl, ok := web.templates[name]
	if !ok {
		web.errorHandler(w, r, fmt.Sprintf("No template found for name: %s", name))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := tmpl.ExecuteTemplate(w, name+".html", data)
	if err != nil {
		web.errorHandler(w, r, fmt.Sprintf("Unable to Excecute Template %v. Error: %v", name, err))
	}
}

func (*web) glob(dir, pattern string) (m []string, e error) {
	m = []string{}
	fi, err := pkger.Stat(dir)
	if err != nil {
		return
	}
	if !fi.IsDir() {
		return
	}
	d, err := pkger.Open(dir)
	if err != nil {
		return
	}
	defer d.Close()

	names, _ := d.Readdir(-1)

	for _, n := range names {
		matched, err := filepath.Match(pattern, n.Name())
		if err != nil {
			return m, err
		}
		if matched {
			m = append(m, path.Join(dir, n.Name()))
		}
	}
	return
}

// parseFiles is a copy of template.ParseFiles modified to use Pkger
func (web *web) parseFiles(t *template.Template, filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("template: no files named in call to ParseFiles")
	}
	for _, filename := range filenames {
		b, err := web.readFile(filename)
		if err != nil {
			return nil, err
		}
		s := string(b)
		name := filepath.Base(filename)
		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

// ReadFile is a copy of os.ReadFile modified to use Pkger
func (web *web) readFile(filename string) ([]byte, error) {
	f, err := pkger.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// It's a good but not certain bet that FileInfo will tell us exactly how much to
	// read, so let's try it but be prepared for the answer to be wrong.
	var n int64 = bytes.MinRead

	if fi, err := f.Stat(); err == nil {
		// As initial capacity for readAll, use Size + a little extra in case Size
		// is zero, and to avoid another allocation after Read has filled the
		// buffer. The readAll call will read into its allocated internal buffer
		// cheaply. If the size was wrong, we'll either waste some space off the end
		// or reallocate as needed, but in the overwhelmingly common case we'll get
		// it just right.
		if size := fi.Size() + bytes.MinRead; size > n {
			n = size
		}
	}
	return web.readAll(f, n)
}

// readAll is a copy of os.readAll needed because of use of
func (*web) readAll(r io.Reader, capacity int64) (b []byte, err error) {
	var buf bytes.Buffer
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	if int64(int(capacity)) == capacity {
		buf.Grow(int(capacity))
	}
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}
