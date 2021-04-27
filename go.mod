module github.com/soapboxsocial/soapbox

go 1.16

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Timothylock/go-signin-with-apple v0.0.0-20210131195746-828dfdd59ab1
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/alicebob/miniredis v2.5.0+incompatible
	github.com/dghubble/go-twitter v0.0.0-20201011215211-4b180d0cc78d
	github.com/dghubble/oauth1 v0.7.0
	github.com/dukex/mixpanel v0.0.0-20180925151559-f8d5594f958e
	github.com/elastic/go-elasticsearch/v7 v7.12.0
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/go-redis/redis/v8 v8.8.2
	github.com/golang/mock v1.5.0
	github.com/gomodule/redigo v1.8.4 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/lib/pq v1.10.1
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pelletier/go-toml v1.9.0 // indirect
	github.com/pion/ice/v2 v2.1.7 // indirect
	github.com/pion/ion-log v1.0.1 // indirect
	github.com/pion/ion-sfu v1.9.8
	github.com/pion/rtp v1.6.5 // indirect
	github.com/pion/webrtc/v3 v3.0.26
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.10.0 // indirect
	github.com/prometheus/common v0.21.0 // indirect
	github.com/rs/zerolog v1.21.0 // indirect
	github.com/segmentio/ksuid v1.0.3
	github.com/sendgrid/rest v2.6.3+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.9.0+incompatible
	github.com/sideshow/apns2 v0.20.0
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/tideland/golib v4.24.2+incompatible // indirect
	github.com/tideland/gorest v2.15.5+incompatible // indirect
	github.com/yuin/gopher-lua v0.0.0-20200816102855-ee81675732da // indirect
	go.opentelemetry.io/otel v0.20.0 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/net v0.0.0-20210423184538-5f58ad60dda6 // indirect
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887 // indirect
	google.golang.org/genproto v0.0.0-20210426193834-eac7f76ac494 // indirect
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/ini.v1 v1.62.0 // indirect
)

replace github.com/pion/ion-sfu => github.com/soapboxsocial/ion-sfu v1.8.2-0.20210421110049-5003aac84a77
