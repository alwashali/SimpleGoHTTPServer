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

type serverArgs struct {
	servingPath string
	port        string
}

var serve serverArgs

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
	<h3>Upload to: {{.Path}} </h3>
	<form
	  enctype="multipart/form-data"
	  action="http://{{.IP}}:{{.Port}}/upload"
	  method="post"
	>
	  <input type="file" name="File" />
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

	file, handler, err := r.FormFile("File")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	dst, err := os.Create(fmt.Sprintf("%s\\%s", serve.servingPath, handler.Filename))
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
			"IP":   ip.String(),
			"Path": serve.servingPath,
			"Port": serve.port,
		}
		display(w, varmap)
	case "POST":
		uploadFile(w, r)
	}
}

func usage() {
	fmt.Println("### SimpleGoHTTPServer ###")
	fmt.Println("\nUsage:\nSimpleGoHTTPServer.exe [port] [path] ")
	fmt.Println("\nExample:\nSimpleGoHTTPServer.exe C:\\temp 8080\n")

	fmt.Printf("File Upload >> /upload\n\n")
}

func main() {
	var servingPath string
	if len(os.Args) <= 2 {
		usage()
	} else {
		serve.port = os.Args[1]
		serve.servingPath = os.Args[2]

		fmt.Println(servingPath)
		http.HandleFunc("/upload", uploadHandler)
		fs := http.FileServer(http.Dir(servingPath))
		http.Handle("/", fs)
		fmt.Printf("Serving %s on http://0.0.0.0:%s \n", serve.servingPath, serve.port)
		fmt.Printf("Upload on http://0.0.0.0:%s/upload \n", serve.port)

		err := http.ListenAndServe(fmt.Sprintf(":%s", serve.port), nil)
		if err != nil {
			fmt.Println(err)
		}
	}

}
