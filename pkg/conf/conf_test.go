package conf_test

import (
	"reflect"
	"testing"

	"github.com/soapboxsocial/soapbox/pkg/conf"
)

func TestLoad(t *testing.T) {
	var conftests = []struct{
		in string
		err bool
		conf *conf.RedisConf
	}{
		{
			"./testdata/redis.toml",
			false,
			&conf.RedisConf{
				Database: 12,
				Port: 1234,
				Password: "test",
				Host: "test",
			},
		},
		{
			"./testdata/invalid.toml",
			true,
			nil,
		},
		{
			"./testdata/wow.toml",
			true,
			nil,
		},
	}

	for _, tt := range conftests {
		t.Run(tt.in, func(t *testing.T) {
			c := &conf.RedisConf{}
			err := conf.Load(tt.in, c)

			if err != nil {
				if tt.err {
					return
				} else {
					t.Fatalf("unexpected err %s", err)
					return
				}
			}


			if !reflect.DeepEqual(c, tt.conf) {
				t.Fatalf("config %v does not match %v", c, tt.conf)
			}
		})
	}
}
