package login

import (
	"database/sql"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/images"
	"github.com/soapboxsocial/soapbox/pkg/login/internal"
	"github.com/soapboxsocial/soapbox/pkg/mail"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

// Contains the login handlers

const expiration = 8760 * time.Hour

const LoginStateRegister = "register"
const LoginStateSuccess = "success"

const TestEmail = "test@apple.com"

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

type appleLoginState struct {
	email string
	user string
}

type Endpoint struct {
	sync.Mutex

	// @todo use redis
	tokens        map[string]tokenState
	registrations map[string]string

	appleLogin map[string]appleLoginState

	users    *users.UserBackend
	sessions *sessions.SessionManager

	ib *images.Backend

	mail *mail.Service

	queue *pubsub.Queue
}

func NewEndpoint(ub *users.UserBackend, manager *sessions.SessionManager, mail *mail.Service, ib *images.Backend, queue *pubsub.Queue) Endpoint {
	return Endpoint{
		tokens:        make(map[string]tokenState),
		registrations: make(map[string]string),
		users:         ub,
		sessions:      manager,
		mail:          mail,
		ib:            ib,
		queue:         queue,
	}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.Path("/start").Methods("POST").HandlerFunc(e.start)
	r.Path("/start/apple").Methods("POST").HandlerFunc(e.loginWithApple)
	r.Path("/pin").Methods("POST").HandlerFunc(e.submitPin)
	r.Path("/register").Methods("POST").HandlerFunc(e.register)

	return r
}

func (e *Endpoint) start(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	email := strings.ToLower(r.Form.Get("email"))
	if !internal.ValidateEmail(email) {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidEmail, "invalid email")
		return
	}

	token, err := internal.GenerateToken()
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	pin, err := internal.GeneratePin()
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	if email == TestEmail {
		pin = "098316"
	}

	e.tokens[token] = tokenState{email: email, pin: pin}

	if email != TestEmail {
		err = e.mail.SendPinEmail(email, pin)
		if err != nil {
			log.Println("failed to send code: ", err.Error())
			httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToLogin, "failed to send code")
		}
	}

	err = json.NewEncoder(w).Encode(map[string]string{"token": token})
	if err != nil {
		log.Println("error writing response: " + err.Error())
	}
}

func (e *Endpoint) loginWithApple(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	email := strings.ToLower(r.Form.Get("email"))
	if !internal.ValidateEmail(email) {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidEmail, "invalid email")
		return
	}

	userID := r.Form.Get("user_id")
	if userID == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid user id")
		return
	}

	token, err := internal.GenerateToken()
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	user, err := e.users.FindByAppleID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			//e.enterRegistrationState(w, token, state.email)
			return
		}

		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	err = e.sessions.NewSession(token, *user, expiration)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	expires := int(expiration.Seconds())
	err = httputil.JsonEncode(w, loginState{State: LoginStateSuccess, User: user, ExpiresIn: &expires})
	if err != nil {
		log.Println("error writing response: " + err.Error())

	}
}

func (e *Endpoint) submitPin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	token := r.Form.Get("token")
	pin := r.Form.Get("pin")

	state, ok := e.tokens[token]
	if !ok {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	if state.pin != pin {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeIncorrectPin, "")
		return
	}

	delete(e.tokens, token)

	user, err := e.users.FindByEmail(state.email)
	if err != nil {
		if err == sql.ErrNoRows {
			e.enterRegistrationState(w, token, state.email)
			return
		}

		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	err = e.sessions.NewSession(token, *user, expiration)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	expires := int(expiration.Seconds())
	err = httputil.JsonEncode(w, loginState{State: LoginStateSuccess, User: user, ExpiresIn: &expires})
	if err != nil {
		log.Println("error writing response: " + err.Error())

	}
}

func (e *Endpoint) enterRegistrationState(w http.ResponseWriter, token, email string) {
	e.registrations[token] = email
	err := httputil.JsonEncode(w, loginState{State: LoginStateRegister})
	if err != nil {
		log.Println("error writing response: " + err.Error())
	}
}

func (e *Endpoint) register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	token := r.Form.Get("token")
	if token == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	email, ok := e.registrations[token]
	if !ok {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	username := strings.ToLower(r.Form.Get("username"))
	if !internal.ValidateUsername(username) {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidUsername, "invalid parameter: username")
		return
	}

	name := strings.TrimSpace(r.Form.Get("display_name"))
	if name == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeMissingParameter, "missing parameter: display_name")
		return
	}

	file, _, err := r.FormFile("profile")
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	image, err := e.processProfilePicture(file)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	// @TODO ALLOW BIO DURING ON-BOARDING
	lastID, err := e.users.CreateUser(email, name, "", image, username)
	if err != nil {
		_ = e.ib.Remove(image)

		if err.Error() == "pq: duplicate key value violates unique constraint \"idx_username\"" {
			httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeUsernameAlreadyExists, "username already exists")
			return
		}

		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToRegister, "failed to register")
		return
	}

	user := users.User{
		ID:          lastID,
		DisplayName: name,
		Username:    username,
		Email:       &email,
		Image:       image,
	}

	err = e.sessions.NewSession(token, user, expiration)
	if err != nil {
		_ = e.ib.Remove(image)

		log.Println("failed to create session: ", err.Error())
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	expires := int(expiration.Seconds())
	err = httputil.JsonEncode(w, loginState{State: LoginStateSuccess, User: &user, ExpiresIn: &expires})
	if err != nil {
		log.Println("error writing response: " + err.Error())
	}

	err = e.queue.Publish(pubsub.UserTopic, pubsub.NewUserEvent(lastID, username))
	if err != nil {
		log.Printf("queue.Publish err: %v\n", err)
	}
}

func (e *Endpoint) processProfilePicture(file multipart.File) (string, error) {
	pngBytes, err := images.MultipartFileToPng(file)
	if err != nil {
		return "", err
	}

	name, err := e.ib.Store(pngBytes)
	if err != nil {
		return "", err
	}

	return name, nil
}
