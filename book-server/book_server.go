package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	StatusCode int
	Msg        string
}

type Book struct {
	Id     int    `json:"Id, omitempty"`
	Title  string `json:"Title, omitempty"`
	Author string `json:"Author, omitempty"`
}

var Port = "8443"
var LoggedIn = false
var srv *http.Server = &http.Server{Addr: Port}
var tlsConfig *tls.Config

var (
	u            = "https://localhost:" + Port
	hello        = "/"
	showBookList = "/showBookList"
	addBook      = "/addBook"
	editBook     = "/editBook/"
	deleteBook   = "/deleteBook/"
	welcome      = "Welcome to the \"Book Server\""
	empty        = "There is no book"
	emptyField   = "contains empty field"
	added        = "added successfully"
	wrongMethod  = "requested method is not allowed"
	wrongId      = "id is required to be an integer"
	edited       = "edited successfully"
	notFound     = "requested book isn't found"
	deleted      = "deleted successfully"
)

var Books []Book

func respond(w http.ResponseWriter, r Response) {
	if r.StatusCode == http.StatusUnauthorized {
		w.Header().Add("WWW-Authenticate", `Basic realm="Authorization Required"`)
	}
	w.WriteHeader(r.StatusCode)
	fmt.Fprintf(w, r.Msg)
}

func checkAuth(r *http.Request) bool {
	// return true
	if !LoggedIn {
		return true
	}

	encodedInfo := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(encodedInfo) != 2 {
		return false
	}

	decodedInfo, err := base64.StdEncoding.DecodeString(encodedInfo[1])

	if err != nil {
		return false
	}

	userInfo := strings.SplitN(string(decodedInfo), ":", 2)
	if len(userInfo) != 2 {
		return false
	}

	if userInfo[0] != "ac" || userInfo[1] != "ac" {
		return false
	}

	return true
}

func Hello(r *http.Request) Response {
	fmt.Println(r.URL, "page may be from clair or any client")
	return Response{http.StatusOK, welcome}
}

func ShowBookList(r *http.Request) Response {
	defer fmt.Println(r.URL, "page")

	if !checkAuth(r) {
		return Response{http.StatusUnauthorized, "unauthorized"}
	}

	if len(Books) == 0 {
		return Response{http.StatusOK, empty}
	}

	list, convertErr := json.MarshalIndent(Books, "", " ")
	if convertErr != nil {
		return Response{http.StatusInternalServerError, "Error occured in converting into json is " + convertErr.Error()}
	}

	return Response{http.StatusOK, string(list)}
}

func AddBook(r *http.Request) Response {
	defer fmt.Println(r.URL, "page")

	if !checkAuth(r) {
		return Response{http.StatusUnauthorized, "unauthorized"}
	}

	var book Book

	if r.Method == "GET" {
		data := r.URL.Query()
		book = Book{Title: data["Title"][0], Author: data["Author"][0]}
	} else if r.Method == "POST" {
		convertErr := json.NewDecoder(r.Body).Decode(&book)
		defer r.Body.Close()

		if convertErr != nil {
			return Response{http.StatusInternalServerError, "error getting json data in PUT method"}
		}
	} else {
		return Response{http.StatusMethodNotAllowed, wrongMethod}
	}

	if book.Title == "" || book.Author == "" {
		return Response{http.StatusBadRequest, emptyField}
	}

	book.Id = len(Books) + 1
	Books = append(Books, book)

	return Response{http.StatusOK, added}
}

func EditBook(r *http.Request) Response {
	defer fmt.Println(r.URL, "page")

	if !checkAuth(r) {
		return Response{http.StatusUnauthorized, "unauthorized"}
	}

	var book Book

	if r.Method == "PUT" {
		id, idErr := strconv.Atoi(r.URL.Path[len(editBook):])

		if idErr != nil {
			return Response{http.StatusBadRequest, wrongId}
		}

		convertErr := json.NewDecoder(r.Body).Decode(&book)
		defer r.Body.Close()

		if convertErr != nil {
			return Response{http.StatusInternalServerError, "error getting json data in PUT method"}
		}

		if book.Title == "" || book.Author == "" {
			return Response{http.StatusBadRequest, emptyField}
		}

		book.Id = id

		for i, _ := range Books {
			if i+1 == id {
				Books[i] = book

				return Response{http.StatusOK, edited}
			}
		}

		return Response{http.StatusBadRequest, notFound}
	} else {
		return Response{http.StatusMethodNotAllowed, wrongMethod}
	}
}

func DeleteBook(r *http.Request) Response {
	defer fmt.Println(r.URL, " page")

	if !checkAuth(r) {
		return Response{http.StatusUnauthorized, "unauthorized"}
	}

	if r.Method == "DELETE" {
		id, idErr := strconv.Atoi(r.URL.Path[len(deleteBook):])

		if idErr != nil {
			return Response{http.StatusBadRequest, wrongId}
		}

		for i, _ := range Books {
			if i+1 == id {
				Books = append(Books[:i], Books[i+1:]...)
				for j, _ := range Books {
					Books[j].Id = j + 1
				}

				return Response{http.StatusOK, deleted}
			}
		}

		return Response{http.StatusBadRequest, notFound}
	} else {
		return Response{http.StatusMethodNotAllowed, wrongMethod}
	}
}

func HandleRequests() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		respond(w, Hello(r))
	})
	http.HandleFunc("/showBookList", func(w http.ResponseWriter, r *http.Request) {
		respond(w, ShowBookList(r))
	})
	http.HandleFunc("/addBook", func(w http.ResponseWriter, r *http.Request) {
		respond(w, AddBook(r))
	})
	http.HandleFunc("/editBook/", func(w http.ResponseWriter, r *http.Request) {
		respond(w, EditBook(r))
	})
	http.HandleFunc("/deleteBook/", func(w http.ResponseWriter, r *http.Request) {
		respond(w, DeleteBook(r))
	})

}

func GetTls(caFile string) {
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig = &tls.Config{
		SessionTicketsDisabled: true,
		ClientCAs:              caCertPool,
		// NoClientCert
		// RequestClientCert
		// RequireAnyClientCert
		// VerifyClientCertIfGiven
		// RequireAndVerifyClientCert
		ClientAuth: tls.NoClientCert, //tls.RequireAndVerifyClientCert,
	}
	tlsConfig.BuildNameToCertificate()
}

func StartServer(certificateFile, privateKeyFile string) {
	srv = &http.Server{
		Addr:         ":" + Port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		//IdleTimeout:  120 * time.Second,
		TLSConfig: tlsConfig,
	}
	fmt.Printf("server-addr-----%s\n", u)
	fmt.Println("Login-----", LoggedIn)
	serverErr := srv.ListenAndServeTLS(certificateFile, privateKeyFile)

	if serverErr != nil {
		log.Fatal("Server Error:", serverErr)
	}
}

func ShutdownServer() {
	if err := srv.Shutdown(context.Background()); err != nil {
		// Error from closing listeners, or context timeout:
		log.Fatalf("HTTP server Shutdown: %v", err)
	}
}

func main() {
	l := len(os.Args)
	fmt.Println("num of args is", l)
	privateKeyFile := os.Args[l-3]
	certificateFile := os.Args[l-2]
	caFile := os.Args[l-1]

	HandleRequests()

	GetTls(caFile)
	StartServer(certificateFile, privateKeyFile)
	// ShutdownServer()
}
