package login

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	httputil "github.com/ephemeral-networks/voicely/pkg/http"
	"github.com/ephemeral-networks/voicely/pkg/sessions"
	"github.com/ephemeral-networks/voicely/pkg/users"
)

// Contains the login handlers

const expiration = 8760 * time.Hour

const sendgrid_api = "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"

const LoginStateRegister = "register"
const LoginStateSuccess = "success"

// @todo better names
type loginState struct {
	State     string      `json:"state"`
	User      *users.User `json:"user,omitempty"`
	ExpiresIn *int        `json:"expires_in,omitempty"`
}

type tokenState struct {
	email string
	pin   string
}

type LoginEndpoint struct {
	sync.Mutex

	// @todo use redis
	tokens        map[string]tokenState
	registrations map[string]string

	users    *users.UserBackend
	sessions *sessions.SessionManager
}

func NewLoginEndpoint(ub *users.UserBackend, manager *sessions.SessionManager) LoginEndpoint {
	return LoginEndpoint{
		tokens:        make(map[string]tokenState),
		registrations: make(map[string]string),
		users:         ub,
		sessions:      manager,
	}
}

func (l *LoginEndpoint) Start(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	email := strings.ToLower(r.Form.Get("email"))
	if !validateEmail(email) {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidEmail, "invalid email")
		return
	}

	token := generateToken()
	pin := generatePin()

	l.tokens[token] = tokenState{email: email, pin: pin}

	err = sendEmail(email, pin)
	if err != nil {
		log.Println("failed to send code: ", err.Error())
		httputil.JsonError(w, 500, httputil.ErrorCodeFailedToLogin, "failed to send code")
	}

	err = json.NewEncoder(w).Encode(map[string]string{"token": token})
	if err != nil {
		log.Println("error writing response: " + err.Error())
	}
}

func (l *LoginEndpoint) SubmitPin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	token := r.Form.Get("token")
	pin := r.Form.Get("pin")

	state, ok := l.tokens[token]
	if !ok {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	// @todo do a specific error
	if state.pin != pin {
		httputil.JsonError(w, 500, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	delete(l.tokens, token)

	user, err := l.users.FindByEmail(state.email)
	if err != nil {
		if err == sql.ErrNoRows {
			l.enterRegistrationState(w, token, state.email)
			return
		}

		httputil.JsonError(w, 500, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	err = l.sessions.NewSession(token, *user, expiration)
	if err != nil {
		httputil.JsonError(w, 500, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	expires := int(expiration.Seconds())
	err = httputil.JsonEncode(w, loginState{State: LoginStateSuccess, User: user, ExpiresIn: &expires})
	if err != nil {
		log.Println("error writing response: " + err.Error())

	}
}

func (l *LoginEndpoint) enterRegistrationState(w http.ResponseWriter, token string, email string) {
	l.registrations[token] = email
	err := httputil.JsonEncode(w, loginState{State: LoginStateRegister})
	if err != nil {
		log.Println("error writing response: " + err.Error())
	}
}

func (l *LoginEndpoint) Register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	token := r.Form.Get("token")
	if token == "" {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	email, ok := l.registrations[token]
	if !ok {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	username := r.Form.Get("username")
	if !validateUsername(username) {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidUsername, "invalid parameter: username")
		return
	}

	name := r.Form.Get("display_name")
	if name == "" {
		httputil.JsonError(w, 400, httputil.ErrorCodeMissingParameter, "missing parameter: display_name")
		return
	}

	lastID, err := l.users.CreateUser(email, name, username)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"idx_username\"" {
			httputil.JsonError(w, 400, httputil.ErrorCodeUsernameAlreadyExists, "username already exists")
			return
		}

		httputil.JsonError(w, 500, httputil.ErrorCodeFailedToRegister, "failed to register")
		return
	}

	user := users.User{ID: lastID, DisplayName: name, Username: username, Email: email}

	err = l.sessions.NewSession(token, user, expiration)
	if err != nil {
		log.Println("failed to create session: ", err.Error())
		httputil.JsonError(w, 500, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	expires := int(expiration.Seconds())
	err = httputil.JsonEncode(w, loginState{State: LoginStateSuccess, User: &user, ExpiresIn: &expires})
	if err != nil {
		log.Println("error writing response: " + err.Error())
	}
}

func sendEmail(recipient string, pin string) error {
	from := mail.NewEmail("Voicely", "no-reply@spksy.app")
	subject := "Voicely Pin"
	to := mail.NewEmail("", recipient)
	plainTextContent := "Your login pin is: " + pin
	htmlContent := fmt.Sprintf("Your login pin is: <strong>%s</strong>", pin)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(sendgrid_api)
	_, err := client.Send(message)
	if err != nil {
		return err
	}

	return nil
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func validateEmail(email string) bool {
	return len(email) < 254 && emailRegex.MatchString(email)
}

var usernameRegex = regexp.MustCompile("^([A-Za-z0-9_]+)*$")

func validateUsername(username string) bool {
	return len(username) < 100 && len(username) > 2 && usernameRegex.MatchString(username)
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
