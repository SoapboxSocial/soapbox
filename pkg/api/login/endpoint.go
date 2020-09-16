package login

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/images"
	"github.com/soapboxsocial/soapbox/pkg/indexer"
	"github.com/soapboxsocial/soapbox/pkg/mail"
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
	email string `json:"email"`
	pin   string `json:"pin"`
}

type LoginEndpoint struct {
	sync.Mutex

	rdb *redis.Client

	// @todo use redis
	//tokens        map[string]tokenState
	registrations map[string]string

	users    *users.UserBackend
	sessions *sessions.SessionManager

	ib *images.Backend

	mail *mail.Service

	index *indexer.Queue
}

func NewLoginEndpoint(ub *users.UserBackend, manager *sessions.SessionManager, mail *mail.Service, ib *images.Backend, index *indexer.Queue, rdb *redis.Client) LoginEndpoint {
	return LoginEndpoint{
		registrations: make(map[string]string),
		users:         ub,
		sessions:      manager,
		mail:          mail,
		ib:            ib,
		index:         index,
		rdb:           rdb,
	}
}

func (l *LoginEndpoint) Start(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	email := strings.ToLower(r.Form.Get("email"))
	if !validateEmail(email) {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidEmail, "invalid email")
		return
	}

	token := generateToken()

	pin := generatePin()
	if email == TestEmail {
		pin = "123456"
	}

	if err := l.setToken(token, tokenState{email: email, pin: pin}); err != nil {
		log.Println("failed to store token: ", err.Error())
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToStoreDevice, "failed to generate login token") // TODO check err message
	}

	if email != TestEmail {
		err = l.mail.SendPinEmail(email, pin)
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

func (l *LoginEndpoint) SubmitPin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	token := r.Form.Get("token")
	pin := r.Form.Get("pin")

	state, err := l.getToken(token)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	if state.pin != pin {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeIncorrectPin, "")
		return
	}

	user, err := l.users.FindByEmail(state.email)
	if err != nil {
		if err == sql.ErrNoRows {
			l.enterRegistrationState(w, token, state.email)
			return
		}

		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	err = l.sessions.NewSession(token, *user, expiration)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	expires := int(expiration.Seconds())
	err = httputil.JsonEncode(w, loginState{State: LoginStateSuccess, User: user, ExpiresIn: &expires})
	if err != nil {
		log.Println("error writing response: " + err.Error())

	}

	if err := l.deleteToken(token); err != nil {
		log.Println("error deleting token: ", err.Error())
	}
}

func (l *LoginEndpoint) enterRegistrationState(w http.ResponseWriter, token, email string) {
	l.registrations[token] = email
	err := httputil.JsonEncode(w, loginState{State: LoginStateRegister})
	if err != nil {
		log.Println("error writing response: " + err.Error())
	}
}

func (l *LoginEndpoint) Register(w http.ResponseWriter, r *http.Request) {
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

	email, ok := l.registrations[token]
	if !ok {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	username := strings.ToLower(r.Form.Get("username"))
	if !validateUsername(username) {
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

	image, err := l.processProfilePicture(file)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	lastID, err := l.users.CreateUser(email, name, image, username)
	if err != nil {
		_ = l.ib.Remove(image)

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

	err = l.sessions.NewSession(token, user, expiration)
	if err != nil {
		_ = l.ib.Remove(image)

		log.Println("failed to create session: ", err.Error())
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToLogin, "")
		return
	}

	expires := int(expiration.Seconds())
	err = httputil.JsonEncode(w, loginState{State: LoginStateSuccess, User: &user, ExpiresIn: &expires})
	if err != nil {
		log.Println("error writing response: " + err.Error())
	}

	l.index.Push(indexer.Event{
		Type:   indexer.EventTypeUserUpdate,
		Params: map[string]interface{}{"id": lastID},
	})
}

func (l *LoginEndpoint) processProfilePicture(file multipart.File) (string, error) {
	imgBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	pngBytes, err := images.ToPNG(imgBytes)
	if err != nil {
		return "", err
	}

	name, err := l.ib.Store(pngBytes)
	if err != nil {
		return "", err
	}

	return name, nil
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func validateEmail(email string) bool {
	return len(email) < 254 && emailRegex.MatchString(email)
}

var usernameRegex = regexp.MustCompile("^([a-z0-9_]+)*$")

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

func (l *LoginEndpoint) setToken(token string, state tokenState) error {
	marshaledState, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return l.rdb.Set(context.Background(), token, marshaledState, time.Minute*10).Err()
}

// TODO CLEAN UP AND DOCUMENT
func (l *LoginEndpoint) getToken(token string) (tokenState, error) {
	rawState, err := l.rdb.Get(context.Background(), token).Bytes()
	if err != nil {
		log.Println("error getting token state: ", err) // TODO err
		return tokenState{}, err
	}

	var state tokenState
	if err := json.Unmarshal(rawState, &state); err != nil {
		log.Println("error unmarshaling token state: ", err) // TODO err
		return tokenState{}, err
	}

	// TODO SOMETHING IS WRONG WITH TOKEN STATE it's not unmarshaling

	return state, nil
}

func (l *LoginEndpoint) deleteToken(token string) error {
	return l.rdb.Del(context.Background(), token).Err()
}

var table = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
