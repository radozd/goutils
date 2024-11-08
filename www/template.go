package www

import (
	"bufio"
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strconv"

	"github.com/radozd/goutils/times"
)

// when used in template like this:
//
//	<link rel="stylesheet" href="{{AutoVersion "main.css"}}">
//
// it generates the following:
//
//	<link rel="stylesheet" href="main.css?1hrtrun">
func tmplAutoVersion(root string, fname string) template.HTML {
	fi, err := os.Stat(filepath.Join(root, fname))
	if err != nil {
		panic(err)
	}
	tm := times.GetTimespec(fi).ModTime()
	return template.HTML(fname + "?" + strconv.FormatInt(tm.Unix(), 32))
}

// when used in template like this:
//
//	{{BuildTime}}
//
// it generates the following:
//
//	2023-01-12 21:34:45
func tmplBuildTime() template.HTML {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	tm := times.GetTimespec(fi).ModTime()
	return template.HTML(tm.Format("2006-01-02 15:04:05"))
}

// template itself must be in `root` folder and has `.tmpl` extension
// functions:
// AutoVersion(fname) adds ?mod_time. files should be relative to `root`
// BuildTime main executable mod time.
func MakeStaticHtml(root string, tmpl string, customFunctions template.FuncMap) {
	funcs := template.FuncMap{
		"AutoVersion": func(fname string) template.HTML { return tmplAutoVersion(root, fname) },
		"BuildTime":   func() template.HTML { return tmplBuildTime() },
	}
	for key, f := range customFunctions {
		funcs[key] = f
	}

	templates := template.Must(template.New("").Funcs(funcs).ParseFiles(filepath.Join(root, tmpl+".tmpl")))

	var processed bytes.Buffer
	templates.ExecuteTemplate(&processed, tmpl, nil)

	f, _ := os.Create(filepath.Join(root, tmpl+".html"))
	w := bufio.NewWriter(f)
	w.WriteString(processed.String())
	w.Flush()
}
