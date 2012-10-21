package myimg

import (
	"html/template"
)

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
	<div><a href="/serve/?blobKey={{.PicKey}}"><img src="/serve/?blobKey={{.PicKey}}" alt="{{.PicKey}}"/></a></div>
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
