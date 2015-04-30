package config

import (
	"github.com/BurntSushi/toml"
	"testing"
	"time"
)

func TestDecodeToml(t *testing.T) {
	var conf TomlConfig
	if _, err := toml.DecodeFile("config.toml", &conf); err != nil {
		t.Fatal(err)
	}
	t.Log(conf.DB)
	t.Log(conf.Device)
	t.Log(conf.Log)
	t.Log(conf.Server)
	t.Log(conf.Shade)
	if conf.Shade.Interval.Duration != time.Second*3+time.Millisecond*6 {
		t.Fatal("Expect %s %s", conf.Shade.Interval, time.Second*3+time.Millisecond*6)
	}
}
