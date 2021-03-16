module github.com/soapboxsocial/soapbox

go 1.14

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Timothylock/go-signin-with-apple v0.0.0-20210131195746-828dfdd59ab1
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/alicebob/miniredis v2.5.0+incompatible
	github.com/dghubble/go-twitter v0.0.0-20201011215211-4b180d0cc78d
	github.com/dghubble/oauth1 v0.7.0
	github.com/dukex/mixpanel v0.0.0-20180925151559-f8d5594f958e
	github.com/elastic/go-elasticsearch/v7 v7.11.0
	github.com/go-redis/redis/v8 v8.7.1
	github.com/golang/mock v1.5.0
	github.com/golang/protobuf v1.4.3
	github.com/gomodule/redigo v1.8.3 // indirect
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/lib/pq v1.10.0
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pion/ion-sfu v1.9.3
	github.com/pion/webrtc/v3 v3.0.14
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.18.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/segmentio/ksuid v1.0.3
	github.com/sendgrid/rest v2.6.2+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.8.0+incompatible
	github.com/sideshow/apns2 v0.20.0
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.1
	github.com/tideland/golib v4.24.2+incompatible // indirect
	github.com/tideland/gorest v2.15.5+incompatible // indirect
	github.com/yuin/gopher-lua v0.0.0-20200816102855-ee81675732da // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/sys v0.0.0-20210309074719-68d13333faf2 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/genproto v0.0.0-20210310155132-4ce2db91004e // indirect
	google.golang.org/grpc v1.36.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/pion/ion-sfu => github.com/soapboxsocial/ion-sfu v1.8.2-0.20210315133713-313e2cf9beaa
