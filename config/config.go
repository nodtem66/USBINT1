// Configuration Type for TOML format
package config

import "time"

type TomlConfig struct {
	Device Device
	DB     Database `toml:"database"`
	Server Server
	Log    Loginfo  `toml:"log"`
	Shade  Shading  `toml:"shading"`
	Sync   SyncInfo `toml:"sync"`
}

type Device struct {
	Name string `toml:"name"`
	Org  string `toml:"organization"`
	Desc string `toml:"description"`
}
type Database struct {
	Path string `toml:"path"`
}

type Server struct {
	Address string `toml:"address"`
}
type Loginfo struct {
	FileName string `toml:"file"`
}
type Shading struct {
	Interval    duration `toml:"interval"`
	MinimumSync int      `toml:"sync"`
}
type SyncInfo struct {
	Interval duration `toml:"interval"`
	DSN      string   `toml:"mysql_dsn"`
}
type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
