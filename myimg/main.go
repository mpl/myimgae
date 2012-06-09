package myimg

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"bytes"
	"fmt"
	"io"
	"crypto/sha1"
	"html/template"
	"net/http"
	"path"
	"time"
)

type Pic struct {
	Author  string
	Name	string
	Content []byte
	Date    time.Time
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/login", login)
	http.HandleFunc("/pic/", pic)
	http.HandleFunc("/upload", upload)
}

var rootTemplate = template.Must(template.New("root").Parse(rootHTML))

const rootHTML = `
<!DOCTYPE html>
<html>
<head>
	<title>ImgDump</title>
</head>
<body>
	<a href="/login">login</a>
	<form action="/upload" method="post" enctype="multipart/form-data">
	<div><input type="file" name="file"></div>
	<div><input type="submit" value="upload"></div>
    </form>
</body>
</html>
`

var picTemplate = template.Must(template.New("pic").Parse(picHTML))

const picHTML = `
<!DOCTYPE html>
<html>
<head>
	<title>ImgDump</title>
</head>
<body>
	<div><a href="/login">login</a></div>

	<div><img src="" alt="{{.Name}}"/></div>

	<form action="/upload" method="post" enctype="multipart/form-data">
	<div><input type="file" name="file"></div>
	<div><input type="submit" value="upload"></div>
    </form>
</body>
</html>
`

func root(w http.ResponseWriter, r *http.Request) {
	if err := rootTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
   if u == nil {
        url, err := user.LoginURL(c, r.URL.String())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Location", url)
        w.WriteHeader(http.StatusFound)
        return
    }
	http.Redirect(w, r, "/", http.StatusFound)
}

func pic(w http.ResponseWriter, r *http.Request) {
	u := r.URL.String()
	_, picName := path.Split(u)
	c := appengine.NewContext(r)
	k := datastore.NewKey(c, "Pic", picName, 0, nil)
    p := Pic{}
    if err := datastore.Get(c, k, &p); err != nil {
		http.Error(w, "Getting from the datastore: " + err.Error(), http.StatusInternalServerError)
        return
    }
	if err := picTemplate.Execute(w, p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hashName := ""
	buf := bytes.NewBuffer(make([]byte, 0))
	// TODO(mpl): limit size
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "reading body: " + err.Error(), http.StatusInternalServerError)
			return
		}
		fileName := part.FileName()
		if fileName == "" {
			continue
		}
		_, err = io.Copy(buf, part)
		if err != nil {
			http.Error(w, "copying: " + err.Error(), http.StatusInternalServerError)
			return
		}
		h := sha1.New()
		_, err = io.Copy(h, buf)
		hashName = fmt.Sprintf("%x", h.Sum(nil))
	}
	c := appengine.NewContext(r)
	pic := Pic{
		Name: hashName,
		Content: buf.Bytes(),
		Date:    time.Now(),
	}
	if u := user.Current(c); u != nil {
		pic.Author = u.String()
	}
	_, err = datastore.Put(c, datastore.NewKey(c, "Pic", hashName, 0, nil), &pic)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/pic/"+hashName, http.StatusFound)
}

/*
package blobstore_example

import (
        "html/template"
        "io"
        "net/http"

        "appengine"
        "appengine/blobstore"
)

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
        w.WriteHeader(http.StatusInternalServerError)
        w.Header().Set("Content-Type", "text/plain")
        io.WriteString(w, "Internal Server Error")
        c.Errorf("%v", err)
}

var rootTemplate = template.Must(template.New("root").Parse(rootTemplateHTML))

const rootTemplateHTML = `
<html><body>
<form action="{{.}}" method="POST" enctype="multipart/form-data">
Upload File: <input type="file" name="file"><br>
<input type="submit" name="submit" value="Submit">
</form></body></html>
`

func handleRoot(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        uploadURL, err := blobstore.UploadURL(c, "/upload", nil)
        if err != nil {
                serveError(c, w, err)
                return
        }
        w.Header().Set("Content-Type", "text/html")
        err = rootTemplate.Execute(w, uploadURL)
        if err != nil {
                c.Errorf("%v", err)
        }
}

func handleServe(w http.ResponseWriter, r *http.Request) {
        blobstore.Send(w, appengine.BlobKey(r.FormValue("blobKey")))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        blobs, _, err := blobstore.ParseUpload(r)
        if err != nil {
                serveError(c, w, err)
                return
        }
        file := blobs["file"]
        if len(file) == 0 {
                c.Errorf("no file uploaded")
                http.Redirect(w, r, "/", http.StatusFound)
                return
        }
        http.Redirect(w, r, "/serve/?blobKey="+string(file[0].BlobKey), http.StatusFound)
}

func init() {
        http.HandleFunc("/", handleRoot)
        http.HandleFunc("/serve/", handleServe)
        http.HandleFunc("/upload", handleUpload)
}
*/