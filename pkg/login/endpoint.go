package login

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/soapboxsocial/soapbox/pkg/apple"
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
	Token     *string     `json:"token,omitempty"`
}

type Endpoint struct {
	sync.Mutex

	state    *StateManager
	users    *users.UserBackend
	sessions *sessions.SessionManager

	ib *images.Backend

	mail *mail.Service

	queue *pubsub.Queue

	signInWithApple apple.SignInWithApple
}

func NewEndpoint(
	ub *users.UserBackend,
	state *StateManager,
	manager *sessions.SessionManager,
	mail *mail.Service,
	ib *images.Backend,
	queue *pubsub.Queue,
	signInWithApple apple.SignInWithApple,
) Endpoint {
	return Endpoint{
		users:           ub,
		state:           state,
		sessions:        manager,
		mail:            mail,
		ib:              ib,
		queue:           queue,
		signInWithApple: signInWithApple,
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

	isApple, err := e.users.IsAppleIDAccount(email)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	if isApple {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid authentication method")
		return
	}

	err = e.state.SetPinState(token, email, pin)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

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

	jwt := r.Form.Get("token")
	if jwt == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid token id")
		return
	}

	userInfo, err := e.signInWithApple.Validate(jwt)
	if err != nil {
		log.Printf("apple validation err: %v", err)
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "failed to validate")
		return
	}

	token, err := internal.GenerateToken()
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	user, err := e.users.FindByAppleID(userInfo.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			e.enterAppleRegistrationState(w, token, userInfo.Email, userInfo.ID)
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
	err = httputil.JsonEncode(w, loginState{State: LoginStateSuccess, User: user, ExpiresIn: &expires, Token: &token})
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

	state, err := e.state.GetState(token)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	if state.Pin != pin {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeIncorrectPin, "")
		return
	}

	e.state.RemoveState(token)

	user, err := e.users.FindByEmail(state.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			e.enterRegistrationState(w, token, state.Email)
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
	err := e.state.SetRegistrationState(token, email)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	err = httputil.JsonEncode(w, loginState{State: LoginStateRegister})
	if err != nil {
		log.Println("error writing response: " + err.Error())
	}
}
func (e *Endpoint) enterAppleRegistrationState(w http.ResponseWriter, token, email, userID string) {
	_, err := e.users.FindByEmail(email)
	if err == nil { // @TODO THIS MEANS THE USER IS ALREADY EXISTING
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid login method for user")
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	err = e.state.SetAppleRegistrationState(token, email, userID)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	err = httputil.JsonEncode(w, loginState{State: LoginStateRegister, Token: &token})
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

	state, err := e.state.GetState(token)
	if err != nil || state.Pin != "" {
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

	var lastID int
	if state.AppleUserID != "" {
		lastID, err = e.users.CreateUserWithAppleLogin(state.Email, name, "", image, username, state.AppleUserID)
	} else {
		lastID, err = e.users.CreateUser(state.Email, name, "", image, username)
	}

	// @TODO ALLOW BIO DURING ON-BOARDING
	if err != nil {
		_ = e.ib.Remove(image)

		if err.Error() == "pq: duplicate key value violates unique constraint \"idx_username\"" {
			httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeUsernameAlreadyExists, "username already exists")
			return
		}

		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToRegister, "failed to register")
		return
	}

	e.state.RemoveState(token)

	user := users.User{
		ID:          lastID,
		DisplayName: name,
		Username:    username,
		Email:       &state.Email,
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
