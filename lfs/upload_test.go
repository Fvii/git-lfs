package lfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestSuccessfulUploadWithVerify(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false
	putCalled := false
	verifyCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != mediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
		err := json.NewDecoder(tee).Decode(reqObj)
		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", buf.String())
		if err != nil {
			t.Fatal(err)
		}

		if reqObj.Oid != "oid" {
			t.Errorf("invalid oid from request: %s", reqObj.Oid)
		}

		if reqObj.Size != 4 {
			t.Errorf("invalid size from request: %d", reqObj.Size)
		}

		obj := &objectResource{
			Links: map[string]*linkRelation{
				"upload": &linkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &linkRelation{
					Href:   server.URL + "/verify",
					Header: map[string]string{"B": "2"},
				},
			},
		}

		by, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}

		postCalled = true
		head := w.Header()
		head.Set("Content-Type", mediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(200)
		w.Write(by)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "PUT" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("A") != "1" {
			t.Error("Invalid A")
		}

		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Error("Invalid Content-Type")
		}

		by, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}

		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", string(by))

		if str := string(by); str != "test" {
			t.Errorf("unexpected body: %s", str)
		}

		putCalled = true
		w.WriteHeader(200)
	})

	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("B") != "2" {
			t.Error("Invalid B")
		}

		if r.Header.Get("Content-Type") != mediaType {
			t.Error("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
		err := json.NewDecoder(tee).Decode(reqObj)
		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", buf.String())
		if err != nil {
			t.Fatal(err)
		}

		if reqObj.Oid != "oid" {
			t.Errorf("invalid oid from request: %s", reqObj.Oid)
		}

		if reqObj.Size != 4 {
			t.Errorf("invalid size from request: %d", reqObj.Size)
		}

		verifyCalled = true
		w.WriteHeader(200)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")

	oidPath := filepath.Join(tmp, "oid")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	wErr := Upload(oidPath, "", nil)
	if wErr != nil {
		t.Fatal(wErr)
	}

	if !postCalled {
		t.Errorf("POST not called")
	}

	if !putCalled {
		t.Errorf("PUT not called")
	}

	if !verifyCalled {
		t.Errorf("verify not called")
	}
}

func TestSuccessfulUploadWithoutVerify(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false
	putCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != mediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
		err := json.NewDecoder(tee).Decode(reqObj)
		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", buf.String())
		if err != nil {
			t.Fatal(err)
		}

		if reqObj.Oid != "oid" {
			t.Errorf("invalid oid from request: %s", reqObj.Oid)
		}

		if reqObj.Size != 4 {
			t.Errorf("invalid size from request: %d", reqObj.Size)
		}

		obj := &objectResource{
			Links: map[string]*linkRelation{
				"upload": &linkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
			},
		}

		by, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}

		postCalled = true
		head := w.Header()
		head.Set("Content-Type", mediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(200)
		w.Write(by)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "PUT" {
			w.WriteHeader(405)
			return
		}

		if a := r.Header.Get("A"); a != "1" {
			t.Errorf("Invalid A: %s", a)
		}

		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Error("Invalid Content-Type")
		}

		by, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}

		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", string(by))

		if str := string(by); str != "test" {
			t.Errorf("unexpected body: %s", str)
		}

		putCalled = true
		w.WriteHeader(200)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")

	oidPath := filepath.Join(tmp, "oid")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	wErr := Upload(oidPath, "", nil)
	if wErr != nil {
		t.Fatal(wErr)
	}

	if !postCalled {
		t.Errorf("POST not called")
	}

	if !putCalled {
		t.Errorf("PUT not called")
	}
}

func TestUploadApiError(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		postCalled = true
		w.WriteHeader(404)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")

	oidPath := filepath.Join(tmp, "oid")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	wErr := Upload(oidPath, "", nil)
	if wErr == nil {
		t.Fatal("no error?")
	}

	if wErr.Panic {
		t.Fatal("should not panic")
	}

	if wErr.Error() != fmt.Sprintf(defaultErrors[404], server.URL+"/media/objects") {
		t.Fatalf("Unexpected error: %s", wErr.Error())
	}

	if !postCalled {
		t.Errorf("POST not called")
	}
}

func TestUploadStorageError(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false
	putCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != mediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
		err := json.NewDecoder(tee).Decode(reqObj)
		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", buf.String())
		if err != nil {
			t.Fatal(err)
		}

		if reqObj.Oid != "oid" {
			t.Errorf("invalid oid from request: %s", reqObj.Oid)
		}

		if reqObj.Size != 4 {
			t.Errorf("invalid size from request: %d", reqObj.Size)
		}

		obj := &objectResource{
			Links: map[string]*linkRelation{
				"upload": &linkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &linkRelation{
					Href:   server.URL + "/verify",
					Header: map[string]string{"B": "2"},
				},
			},
		}

		by, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}

		postCalled = true
		head := w.Header()
		head.Set("Content-Type", mediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(200)
		w.Write(by)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		putCalled = true
		w.WriteHeader(404)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")

	oidPath := filepath.Join(tmp, "oid")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	wErr := Upload(oidPath, "", nil)
	if wErr == nil {
		t.Fatal("no error?")
	}

	if wErr.Panic {
		t.Fatal("should not panic")
	}

	if wErr.Error() != fmt.Sprintf(defaultErrors[404], server.URL+"/upload") {
		t.Fatalf("Unexpected error: %s", wErr.Error())
	}

	if !postCalled {
		t.Errorf("POST not called")
	}

	if !putCalled {
		t.Errorf("PUT not called")
	}
}

func TestUploadVerifyError(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	tmp := tempdir(t)
	defer server.Close()
	defer os.RemoveAll(tmp)

	postCalled := false
	putCalled := false
	verifyCalled := false

	mux.HandleFunc("/media/objects", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("Accept") != mediaType {
			t.Errorf("Invalid Accept")
		}

		if r.Header.Get("Content-Type") != mediaType {
			t.Errorf("Invalid Content-Type")
		}

		buf := &bytes.Buffer{}
		tee := io.TeeReader(r.Body, buf)
		reqObj := &objectResource{}
		err := json.NewDecoder(tee).Decode(reqObj)
		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", buf.String())
		if err != nil {
			t.Fatal(err)
		}

		if reqObj.Oid != "oid" {
			t.Errorf("invalid oid from request: %s", reqObj.Oid)
		}

		if reqObj.Size != 4 {
			t.Errorf("invalid size from request: %d", reqObj.Size)
		}

		obj := &objectResource{
			Links: map[string]*linkRelation{
				"upload": &linkRelation{
					Href:   server.URL + "/upload",
					Header: map[string]string{"A": "1"},
				},
				"verify": &linkRelation{
					Href:   server.URL + "/verify",
					Header: map[string]string{"B": "2"},
				},
			},
		}

		by, err := json.Marshal(obj)
		if err != nil {
			t.Fatal(err)
		}

		postCalled = true
		head := w.Header()
		head.Set("Content-Type", mediaType)
		head.Set("Content-Length", strconv.Itoa(len(by)))
		w.WriteHeader(200)
		w.Write(by)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Server: %s %s", r.Method, r.URL)

		if r.Method != "PUT" {
			w.WriteHeader(405)
			return
		}

		if r.Header.Get("A") != "1" {
			t.Error("Invalid A")
		}

		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Error("Invalid Content-Type")
		}

		by, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}

		t.Logf("request header: %v", r.Header)
		t.Logf("request body: %s", string(by))

		if str := string(by); str != "test" {
			t.Errorf("unexpected body: %s", str)
		}

		putCalled = true
		w.WriteHeader(200)
	})

	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		verifyCalled = true
		w.WriteHeader(404)
	})

	Config.SetConfig("lfs.url", server.URL+"/media")

	oidPath := filepath.Join(tmp, "oid")
	if err := ioutil.WriteFile(oidPath, []byte("test"), 0744); err != nil {
		t.Fatal(err)
	}

	wErr := Upload(oidPath, "", nil)
	if wErr == nil {
		t.Fatal("no error?")
	}

	if wErr.Panic {
		t.Fatal("should not panic")
	}

	if wErr.Error() != fmt.Sprintf(defaultErrors[404], server.URL+"/verify") {
		t.Fatalf("Unexpected error: %s", wErr.Error())
	}

	if !postCalled {
		t.Errorf("POST not called")
	}

	if !putCalled {
		t.Errorf("PUT not called")
	}

	if !verifyCalled {
		t.Errorf("verify not called")
	}
}
