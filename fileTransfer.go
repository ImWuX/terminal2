package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func (ctx *Context) fileDownload(w http.ResponseWriter, r *http.Request) {
	var conn *Connection
	if conn = ctx.getConnectionLock(r.URL.Query().Get("connectionId")); conn == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid connection id"))
		return
	}

	path := conn.downloadPath
	conn.downloadPath = ""

	if !filepath.IsAbs(path) {
		proc, err := ctx.procFs.Proc(conn.proc.Pid)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to query proc"))
			return
		}
		cwd, err := proc.Cwd()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to query CWD"))
			return
		}
		path = filepath.Join(cwd, path)
	}

	file, err := os.Open(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to read requested file"))
		return
	}

	w.Header().Add("content-disposition", fmt.Sprintf("attachment; filename=%s;", filepath.Base(path)))
	io.Copy(w, file)
}

func (ctx *Context) fileUpload(w http.ResponseWriter, r *http.Request) {
	var conn *Connection
	if conn = ctx.getConnectionLock(r.URL.Query().Get("connectionId")); conn == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid connection id"))
		return
	}

	// Parse our multipart form, 10 << 20 specifies a maximum upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to parse form"))
		return
	}
	defer file.Close()

	proc, err := ctx.procFs.Proc(conn.proc.Pid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to query proc"))
		return
	}
	cwd, err := proc.Cwd()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to query CWD"))
		return
	}

	var name string
	counter := 0
	for {
		name = handler.Filename
		if counter > 0 {
			name = fmt.Sprintf("%s-%d", name, counter)
		}
		if !FileExists(filepath.Join(cwd, name)) {
			break
		}
		counter++
	}
	destFile, err := os.Create(filepath.Join(cwd, name))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to create file for upload"))
		return
	}
	defer destFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to read bytes"))
		return
	}

	_, err = destFile.Write(fileBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to write bytes"))
		return
	}

	w.Write([]byte(fmt.Sprintf("Uploaded file %s to %s", name, cwd)))
}
