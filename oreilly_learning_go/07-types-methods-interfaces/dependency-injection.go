// Pretend web app with various requirements
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// FastLog writes log messages out so fast
func FastLog(message string) {
	fmt.Println(message)
}

// LogAdapter adapts function to Logger interface
type LogAdapter func(string)

func (l LogAdapter) Log(message string) {
	l(message)
}

// FastDataStore is a high-performance in-memory non-SQL database
type FastDataStore struct {
	users map[int]string
}

func NewFastDataStore() DataStore {
	return FastDataStore{
		users: map[int]string{
			1: "leon",
			2: "alyson",
			3: "blake",
			4: "stella",
		},
	}
}

func (s FastDataStore) Username(id int) (string, bool) {
	name, ok := s.users[id]
	return name, ok
}

// Interfaces for business logic; depend on interfaces not concrete types
type DataStore interface {
	Username(id int) (string, bool)
}

type Logger interface {
	Log(message string)
}

// Business logic
type Business struct {
	l Logger
	s DataStore
}

func NewBusiness(l Logger, s DataStore) Business {
	return Business{l, s}
}

func (b Business) SayHello(id int) (string, error) {
	b.l.Log(fmt.Sprintf("in SayHello for %d", id))
	name, ok := b.s.Username(id)
	if !ok {
		return "", fmt.Errorf("unknown id: %d")
	}
	return "Hello, " + name, nil
}

func (b Business) SayGoodbye(id int) (string, error) {
	b.l.Log(fmt.Sprintf("in SayGoodbye for %d", id))
	name, ok := b.s.Username(id)
	if !ok {
		return "", fmt.Errorf("unknown id: %d")
	}
	return "Goodbye, " + name, nil
}

// Controller
type Controller struct {
	l Logger
	b Business
}

func NewController(l Logger, b Business) Controller {
	return Controller{l, b}
}

func (c Controller) SayHello(w http.ResponseWriter, r *http.Request) {
	c.l.Log("Controller.SayHello()")
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	message, err := c.b.SayHello(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(message))
}

func (c Controller) SayGoodbye(w http.ResponseWriter, r *http.Request) {
	c.l.Log("Controller.Goodbye()")
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	message, err := c.b.SayGoodbye(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(message))
}

const addr = ":8000"

// Main
func main() {
	// Concrete dependencies
	l := LogAdapter(FastLog)
	s := NewFastDataStore()

	// Logic / controller
	b := NewBusiness(l, s)
	c := NewController(l, b)

	// HTTP Server
	http.HandleFunc("/hello", c.SayHello)
	http.HandleFunc("/goodbye", c.SayGoodbye)

	l.Log("Running HTTP on " + addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
