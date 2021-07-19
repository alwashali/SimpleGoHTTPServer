package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"text/template"
)

var UploadPage string = `

<!DOCTYPE html>
<html lang="en">
  <head>
	<meta charset="UTF-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<meta http-equiv="X-UA-Compatible" content="ie=edge" />
	<title>Upload File</title>
  </head>
  <body>
	<form
	  enctype="multipart/form-data"
	  action="http://{{.IP}}:8080/upload"
	  method="post"
	>
	  <input type="file" name="myFile" />
	  <input type="submit" value="upload" />
	</form>
  </body>
</html>

`

var t, err = template.New("upload").Parse(UploadPage)

func display(w http.ResponseWriter, data interface{}) {
	err = t.ExecuteTemplate(w, "upload", data)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	dst, err := os.Create(handler.Filename)
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ip := GetOutboundIP()
		varmap := map[string]interface{}{
			"IP": ip.String(),
		}
		display(w, varmap)
	case "POST":
		uploadFile(w, r)
	}
}

func main() {
	servingPath := os.Args[1]
	http.HandleFunc("/upload", uploadHandler)
	fs := http.FileServer(http.Dir(servingPath))
	http.Handle("/", fs)
	fmt.Println("Listening...:0.0.0.0:8080")
	http.ListenAndServe(":8080", nil)
}
