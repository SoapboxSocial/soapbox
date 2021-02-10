package conf_test

import (
	"testing"

	"github.com/soapboxsocial/soapbox/pkg/conf"
)

func TestLoad(t *testing.T) {

	c := &conf.RedisConf{}
	err := conf.Load("./testdata/redis.toml", c)
	if err != nil {
		t.Fatal(err)
	}

	if c.Database != 12 {
		t.Fatalf("unexpected database %d", c.Database)
	}

	if c.Port != 1234 {
		t.Fatalf("unexpected port %d", c.Port)
	}

	if c.Password != "test" {
		t.Fatalf("unexpected password %s", c.Password)
	}

	if c.Host != "test" {
		t.Fatalf("unexpected host %s", c.Host)
	}
}

func TestLoad_FailureWithInvalidConf(t *testing.T) {

	c := &conf.RedisConf{}
	err := conf.Load("./testdata/invalid.toml", c)
	if err == nil {
		t.Fatal("loading did not fail")
	}
}

func TestLoad_FailureWithInvalidPath(t *testing.T) {

	c := &conf.RedisConf{}
	err := conf.Load("./testdata/wow.toml", c)
	if err == nil {
		t.Fatal("loading did not fail")
	}
}
