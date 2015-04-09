package config

// Configuration Type
type TomlConfig struct {
	Device Device
	DB     Database `toml:"database"`
	Server Server
	Log    Loginfo `toml:"log"`
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
