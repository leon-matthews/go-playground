package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

var http500 = "Internal error sent to development team"
var portNumber = 8080

func main() {
	router := http.NewServeMux()
	router.HandleFunc("GET /{$}", index)
	router.HandleFunc("POST /{$}", send)
	router.HandleFunc("GET /thanks/", thanks)
	router.HandleFunc("GET /time/", timeNow)

	// Static files
	fileHandler := http.StripPrefix("/s/", http.FileServer(http.Dir("static")))
	router.Handle("/s/", fileHandler)

	port := fmt.Sprintf(":%d", portNumber)
	fmt.Println("Starting HTTP server on port", port)
	server := &http.Server{
		Addr:    port,
		Handler: router,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

// index renders homepage with empty message form
func index(w http.ResponseWriter, r *http.Request) {
	render(w, "html/index.html", Message{})
}

// send renders form errors OR sends message and shows thanks
func send(w http.ResponseWriter, r *http.Request) {
	// Handle form
	msg := &Message{
		Email:   r.PostFormValue("email"),
		Content: r.PostFormValue("content"),
	}

	if msg.Validate() == false {
		render(w, "html/index.html", msg)
		return
	}

	r.ParseForm()
	log.Println(r.Form)
	log.Println(r.PostForm)

	// Redirect to thanks
	http.Redirect(w, r, "/thanks/", http.StatusSeeOther)
}

func thanks(w http.ResponseWriter, r *http.Request) {
	render(w, "html/thanks.html", nil)
}

func timeNow(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Auckland, right now: " + time.Now().Format(time.UnixDate)))
}

// render writes the output from executing the template to the response
func render(w http.ResponseWriter, filename string, data interface{}) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		log.Print(err)
		http.Error(w, http500, http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Print(err)
		http.Error(w, http500, http.StatusInternalServerError)
	}
}
