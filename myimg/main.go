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
)

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/pic/", pic)
	http.HandleFunc("/serve/", serve)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/pics", allPics)
}

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Internal Server Error")
	c.Errorf("%v", err)
}

func isAnon(c appengine.Context) bool {
	u := user.Current(c)
	return (u == nil) 
}

func login(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
//	c.Debugf(r.URL.Path)
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
	if r.URL.Path == "/login" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
//	c.Debugf(r.URL.Path)
	u := user.Current(c)
	if u != nil {
		url, err := user.LogoutURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}
	if r.URL.Path == "/logout" {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

var rootTemplate = template.Must(template.New("root").Parse(rootHTML))

const rootHTML = `
<!DOCTYPE html>
<html>
<head>
	<title>ImgDump</title>
</head>
<body>
	{{ if .Anon }}<div> Log in to upload or list your previous uploads </div>{{ end }}
	<div> <a href="/">home</a> <a href="/login">login</a> <a href="/logout">logout</a> </div>
	{{ if not .Anon }}
	<div><a href="/pics">list</a></div>
	<form action="{{.UploadURL}}" method="post" enctype="multipart/form-data">
	<div><input type="file" name="file"></div>
	<div><input type="submit" value="upload"></div>
    </form>
	{{ end }}
</body>
</html>
`

type serveRoot struct {
	UploadURL string
	Anon bool
}

func root(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	uploadURL, err := blobstore.UploadURL(c, "/upload", nil)
	if err != nil {
		serveError(c, w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	served := serveRoot{uploadURL.String(), isAnon(c)}
	if err := rootTemplate.Execute(w, served); err != nil {
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
	{{ if .Anon }}<div> Log in to upload or list your previous uploads </div>{{ end }}
	<div> <a href="/">home</a> <a href="/login">login</a> <a href="/logout">logout</a> </div>
	<div><img src="/serve/?blobKey={{.PicKey}}" alt="{{.PicKey}}"/></div>
	{{ if not .Anon }}
	<div><a href="/pics">list</a></div>
	<form action="{{.Upload}}" method="post" enctype="multipart/form-data">
	<div><input type="file" name="file"></div>
	<div><input type="submit" value="upload"></div>
    </form>
	{{ end }}
</body>
</html>
`

type servePic struct {
	Upload string
	PicKey string
	Anon bool
}

type shortToLong struct {
	Owner string
	Hash string
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
	k := datastore.NewKey(c, "shortKey", picName, 0, nil)
	short := shortToLong{}
	if err := datastore.Get(c, k, &short); err != nil {
		http.Error(w, "Getting from the datastore: "+err.Error(), http.StatusInternalServerError)
		return
	}
	//	key := r.FormValue("blobKey")
	p := servePic{uploadURL.String(), short.Hash, isAnon(c)}
	w.Header().Set("Content-Type", "text/html")
	if err := picTemplate.Execute(w, p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		url, err := user.LoginURL(c, "/")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}
	blobs, _, err := blobstore.ParseUpload(r)
	if err != nil {
//		serveError(c, w, err)
		c.Errorf("%v", err)
		// TODO(mpl): probably not a "StatusFound" that we want here
		http.Redirect(w, r, "/", http.StatusFound)
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
	_, err = datastore.Put(c, datastore.NewKey(c, "shortKey", short, 0, nil), &shortToLong{u.String(), long})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//	http.Redirect(w, r, "/pic/?blobKey="+string(file[0].BlobKey), http.StatusFound)
	http.Redirect(w, r, "/pic/"+short, http.StatusFound)
}

var picsTemplate = template.Must(template.New("pics").Parse(picsHTML))

const picsHTML = `
<!DOCTYPE html>
<html>
<head>
	<title>ImgDump</title>
</head>
<body>
	<div> <a href="/">home</a> <a href="/login">login</a> <a href="/logout">logout</a> </div>
	<ul>
	{{range .}} <li> <a href="pic/{{.}}"> {{.}} </li> {{end}}
	</ul>
</body>
</html>
`

func allPics(w http.ResponseWriter, r *http.Request) {
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
//	q := datastore.NewQuery("shortKey").Limit(10)
	q := datastore.NewQuery("shortKey").Filter("Owner =", u.String())
	longs := make([]shortToLong, 0, 10)
	keys, err := q.GetAll(c, &longs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	shorts := make([]string, 0, 1)
	for _,v := range keys {
		shorts = append(shorts, v.StringID())
	}
	
	if err := picsTemplate.Execute(w, shorts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

