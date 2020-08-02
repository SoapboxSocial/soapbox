package login

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/ephemeral-networks/voicely/pkg/sessions"
	"github.com/ephemeral-networks/voicely/pkg/users"
)

// Contains the login handlers

// @todo find a btter name
type tokenState struct {
	email string
	pin string
}

type Login struct {
	sync.Mutex

	tokens map[string]tokenState

	users *users.UserBackend
	sessions *sessions.SessionManager
}

func NewLogin(ub *users.UserBackend, manager *sessions.SessionManager) Login {
	return Login{tokens: make(map[string]tokenState), users: ub, sessions: manager}
}

func (l *Login) Start(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// @todo
		fmt.Println("fuck")
		return
	}
	email := r.Form.Get("email")
	if email == "" {
		// @todo
		return
	}

	// @todo check that email is set
	token := generateToken()
	pin := generatePin()

	l.tokens[token] = tokenState{email: email, pin: pin}

	// @todo cleanup
	err = json.NewEncoder(w).Encode(map[string]string{"token": token})
	if err != nil {
		fmt.Println(err)
	}

	log.Println("pin:" + pin)
}

func (l *Login) SubmitPin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// @todo
		fmt.Println("fuck")
		return
	}

	token := r.Form.Get("token")
	pin := r.Form.Get("pin")

	state := l.tokens[token]
	if state.pin != pin {
		// @todo send failure
		return
	}

	delete(l.tokens, token)

	fmt.Println(state)

	log.Println("success")

	// @todo make account if not exist

	// @todo start session
}

func (l *Login) Register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// @todo
		fmt.Println("fuck")
		return
	}

	token := r.Form.Get("token")
	if token == "" {
		// @todo
		return
	}

	username := r.Form.Get("username")
	if username == "" {
		// @todo bad request
		return
	}

	displayname := r.Form.Get("display_name")
	if displayname == "" {
		// @todo bad request
		return
	}

	//token := r.Form.Get("token")
}


func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func generatePin() string {
	max := 6
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

var table = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
