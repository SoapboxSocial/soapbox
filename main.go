package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	signinwithapple "github.com/Timothylock/go-signin-with-apple/apple"
	"github.com/dghubble/oauth1"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gorilla/mux"
	"github.com/sendgrid/sendgrid-go"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/account"
	"github.com/soapboxsocial/soapbox/pkg/activeusers"
	"github.com/soapboxsocial/soapbox/pkg/analytics"
	"github.com/soapboxsocial/soapbox/pkg/apple"
	"github.com/soapboxsocial/soapbox/pkg/blocks"
	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/followers"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/http/middlewares"
	"github.com/soapboxsocial/soapbox/pkg/images"
	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
	"github.com/soapboxsocial/soapbox/pkg/login"
	"github.com/soapboxsocial/soapbox/pkg/mail"
	"github.com/soapboxsocial/soapbox/pkg/me"
	"github.com/soapboxsocial/soapbox/pkg/minis"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/search"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/sql"
	"github.com/soapboxsocial/soapbox/pkg/stories"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Conf struct {
	Twitter struct {
		Key    string `mapstructure:"key"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"twitter"`
	Sendgrid struct {
		Key string `mapstructure:"key"`
	} `mapstructure:"sendgrid"`
	CDN struct {
		Images  string `mapstructure:"images"`
		Stories string `mapstructure:"stories"`
	} `mapstructure:"cdn"`
	Apple  conf.AppleConf    `mapstructure:"apple"`
	Redis  conf.RedisConf    `mapstructure:"redis"`
	DB     conf.PostgresConf `mapstructure:"db"`
	GRPC   conf.AddrConf     `mapstructure:"grpc"`
	Listen conf.AddrConf     `mapstructure:"listen"`
	Login  login.Config      `mapstructure:"login"`
}

func parse() (*Conf, error) {
	var file string
	flag.StringVar(&file, "c", "config.toml", "config file")
	flag.Parse()

	config := &Conf{}
	err := conf.Load(file, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func main() {
	config, err := parse()
	if err != nil {
		log.Fatalf("failed to parse config err: %v", err)
	}

	rdb := redis.NewRedis(config.Redis)
	queue := pubsub.NewQueue(rdb)

	db, err := sql.Open(config.DB)
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}

	s := sessions.NewSessionManager(rdb)
	ub := users.NewBackend(db)
	fb := followers.NewFollowersBackend(db)
	ns := notifications.NewStorage(rdb)

	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	devicesBackend := devices.NewBackend(db)

	amw := middlewares.NewAuthenticationMiddleware(s)

	r := mux.NewRouter()

	r.MethodNotAllowedHandler = http.HandlerFunc(httputil.NotAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(httputil.NotFoundHandler)

	ib := images.NewImagesBackend(config.CDN.Images)
	ms := mail.NewMailService(sendgrid.NewSendClient(config.Sendgrid.Key))

	loginState := login.NewStateManager(rdb)

	secret, err := ioutil.ReadFile(config.Apple.Path)
	if err != nil {
		panic(err)
	}

	appleClient, err := apple.NewSignInWithAppleAppValidation(
		signinwithapple.New(),
		config.Apple.TeamID,
		config.Apple.Bundle,
		config.Apple.KeyID,
		string(secret),
	)

	if err != nil {
		panic(err)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", config.GRPC.Host, config.GRPC.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	roomService := pb.NewRoomServiceClient(conn)

	loginEndpoints := login.NewEndpoint(ub, loginState, s, ms, ib, queue, appleClient, roomService, config.Login)
	loginRouter := loginEndpoints.Router()
	mount(r, "/v1/login", loginRouter)

	storiesBackend := stories.NewBackend(db)

	usersEndpoints := users.NewEndpoint(
		ub,
		fb,
		s,
		ib,
		queue,
		storiesBackend,
	)
	usersRouter := usersEndpoints.Router()
	usersRouter.Use(amw.Middleware)
	mount(r, "/v1/users", usersRouter)

	storiesEndpoint := stories.NewEndpoint(storiesBackend, stories.NewFileBackend(config.CDN.Stories), queue)
	storiesRouter := storiesEndpoint.Router()
	storiesRouter.Use(amw.Middleware)
	mount(r, "/v1/stories", storiesRouter)

	devicesEndpoint := devices.NewEndpoint(devicesBackend)
	devicesRoutes := devicesEndpoint.Router()
	devicesRoutes.Use(amw.Middleware)
	mount(r, "/v1/devices", devicesRoutes)

	accountEndpoint := account.NewEndpoint(account.NewBackend(db), queue, s)
	accountRouter := accountEndpoint.Router()
	accountRouter.Use(amw.Middleware)
	mount(r, "/v1/account", accountRouter)

	blocksBackend := blocks.NewBackend(db)
	blocksEndpoint := blocks.NewEndpoint(blocksBackend)
	blocksRouter := blocksEndpoint.Router()
	blocksRouter.Use(amw.Middleware)
	mount(r, "/v1/blocks", blocksRouter)

	// twitter oauth config
	oauth := oauth1.NewConfig(
		config.Twitter.Key,
		config.Twitter.Secret,
	)

	meEndpoint := me.NewEndpoint(me.Parameters{
		Users:                 ub,
		NotificationsStorage:  ns,
		OauthConfig:           oauth,
		LinkedAccountsBackend: linkedaccounts.NewLinkedAccountsBackend(db),
		StoriesBackend:        storiesBackend,
		Queue:                 queue,
		ActiveUsersBackend:    activeusers.NewBackend(db),
		NotificationSettings:  notifications.NewSettings(db),
	})

	meRoutes := meEndpoint.Router()

	meRoutes.Use(amw.Middleware)
	mount(r, "/v1/me", meRoutes)

	searchEndpoint := search.NewEndpoint(client)
	searchRouter := searchEndpoint.Router()
	searchRouter.Use(amw.Middleware)
	mount(r, "/v1/search", searchRouter)

	minisBackend := minis.NewBackend(db)
	minisEndpoint := minis.NewEndpoint(minisBackend)

	minisRouter := minisEndpoint.Router()
	minisRouter.Use(amw.Middleware)
	mount(r, "/v1/minis", minisRouter)

	analyticsBackend := analytics.NewBackend(db)
	analyticsEndpoint := analytics.NewEndpoint(analyticsBackend)
	analyticsRouter := analyticsEndpoint.Router()
	analyticsRouter.Use(amw.Middleware)
	mount(r, "/v1/analytics", analyticsRouter)

	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Listen.Port), httputil.CORS(r))
	if err != nil {
		log.Print(err)
	}
}

func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			AddSlashForRoot(handler),
		),
	)
}

// AddSlashForRoot adds a slash if the path is the root path.
// This is necessary for our subrouters where there may be a root.
func AddSlashForRoot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// @TODO MAYBE ENSURE SUFFIX DOESN'T ALREADY EXIST?
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}

		next.ServeHTTP(w, r)
	})
}
