package myimg

import (
	"appengine"
	"appengine/blobstore"
	"appengine/datastore"
	"appengine/user"
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path"
	"time"
)

func init() {
	http.HandleFunc("/", root)
	// http.HandleFunc("/login", login)
	http.HandleFunc("/pic/", pic)
	http.HandleFunc("/serve/", serve)
	http.HandleFunc("/upload", upload)
}

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Internal Server Error")
	c.Errorf("%v", err)
}

type Pic struct {
	Author  string
	Name    string
	Content []byte
	Date    time.Time
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

var rootTemplate = template.Must(template.New("root").Parse(rootHTML))

const rootHTML = `
<!DOCTYPE html>
<html>
<head>
	<title>ImgDump</title>
</head>
<body>
	<div><a href="/">home</a></div>
<!--
	<a href="/login">login</a>
-->
	<form action="{{.}}" method="post" enctype="multipart/form-data">
	<div><input type="file" name="file"></div>
	<div><input type="submit" value="upload"></div>
    </form>
</body>
</html>
`

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	uploadURL, err := blobstore.UploadURL(c, "/upload", nil)
	if err != nil {
		serveError(c, w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	if err := rootTemplate.Execute(w, uploadURL); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func serve(w http.ResponseWriter, r *http.Request) {
	blobstore.Send(w, appengine.BlobKey(r.FormValue("blobKey")))
}

var picTemplate = template.Must(template.New("pic").Parse(picHTML))

const picHTML = `
<!DOCTYPE html>
<html>
<head>
	<title>ImgDump</title>
</head>
<body>
	<div><a href="/">home</a></div>
<!--
	<div><a href="/login">login</a></div>
-->

	<div><img src="/serve/?blobKey={{.PicKey}}" alt="{{.PicKey}}"/></div>

	<form action="{{.Upload}}" method="post" enctype="multipart/form-data">
	<div><input type="file" name="file"></div>
	<div><input type="submit" value="upload"></div>
    </form>
</body>
</html>
`

type servePic struct {
	Upload string
	PicKey string
}

type shortened struct {
	Long string
}

func pic(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	uploadURL, err := blobstore.UploadURL(c, "/upload", nil)
	if err != nil {
		serveError(c, w, err)
		return
	}
	u := r.URL.String()
	_, picName := path.Split(u)
	k := datastore.NewKey(c, "string", picName, 0, nil)
	short := shortened{}
	if err := datastore.Get(c, k, &short); err != nil {
		http.Error(w, "Getting from the datastore: "+err.Error(), http.StatusInternalServerError)
		return
	}
	//	key := r.FormValue("blobKey")
	p := servePic{uploadURL.String(), short.Long}
	w.Header().Set("Content-Type", "text/html")
	if err := picTemplate.Execute(w, p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
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
	long := string(file[0].BlobKey)
	h := md5.New()
	_, err = io.WriteString(h, long)
	if err != nil {
		serveError(c, w, err)
	}
	short := fmt.Sprintf("%x", h.Sum(nil))
	_, err = datastore.Put(c, datastore.NewKey(c, "string", short, 0, nil), &shortened{long})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//	http.Redirect(w, r, "/pic/?blobKey="+string(file[0].BlobKey), http.StatusFound)
	http.Redirect(w, r, "/pic/"+short, http.StatusFound)
}
