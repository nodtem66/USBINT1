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
	t.Log(conf.Sync)
	if conf.Shade.Interval.Duration != time.Second*10 {
		t.Fatal("Expect %s %s", conf.Shade.Interval, time.Second*10)
	}
	if conf.Sync.Interval.Duration != time.Second*10 {
		t.Fatal("Expect %s %s", conf.Shade.Interval, time.Second*10)
	}
	if conf.Sync.ShadeTime.Duration != time.Second*10 {
		t.Fatal("Expect %s %s", conf.Shade.Interval, time.Second*10)
	}
}
