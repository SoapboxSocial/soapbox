package login

import (
	"crypto/rand"
	"database/sql"
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

const LoginStateRegister = "register"
const LoginStateSuccess = "success"

type loginState struct {
	State string `json:"state"`
	User *users.User `json:"user,omitempty"`
}

type Login struct {
	sync.Mutex

	tokens map[string]tokenState

	registrations map[string]string

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

	user, err := l.users.FindByEmail(state.email)
	if err != nil {
		if err == sql.ErrNoRows {
			l.enterRegistrationState(w, token, state.email)
			return
		}

		// @todo
		return
	}

	l.sessions.NewSession(token, *user)

	err = jsonEncode(w, loginState{State: LoginStateSuccess, User: user})
	if err != nil {
		// @todo
		return
	}
}

func (l *Login) enterRegistrationState(w http.ResponseWriter, token string, email string) {
	l.registrations[token] = email
	err := jsonEncode(w, loginState{State: LoginStateRegister})
	if err != nil {
		// @todo
		return
	}
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

	email, ok := l.registrations[token]
	if !ok {
		// @todo bad request
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

	lastID, err := l.users.CreateUser(email, displayname, username)
	if err != nil {
		// @todo
		return
	}

	user := users.User{ID: lastID, DisplayName: displayname, Username: username, Email: email}

	l.sessions.NewSession(token, user)

	err = jsonEncode(w, user)
	if err != nil {
		// @todo
		return
	}
}

func jsonEncode(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
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
