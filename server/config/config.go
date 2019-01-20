package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

func Load(file string) (*Config, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	err = yaml.Unmarshal(content, conf)
	if err != nil {
		return nil, err
	}
	conf.Path = file
	return conf, nil
}

func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.Path, data, 0644)
}

func Default() *Config {
	return &Config{
		StatusRateSeconds: 3,
		Port:              5000,
		Servers:           make(map[string]*ServerOptions),
	}
}

type Config struct {
	Path string `yaml:"-"`

	ShooterGame       string                    `yaml:"shooterGame"`
	StatusRateSeconds int                       `yaml:"statusRateSeconds"`
	Port              int                       `yaml:"port"`
	Servers           map[string]*ServerOptions `yaml:"servers"`
}

type ServerOptions struct {
	Pid                  int      `yaml:"pid"`
	ServerX              int      `yaml:"serverX"`
	ServerY              int      `yaml:"serverY"`
	Port                 int      `yaml:"port"`
	QueryPort            int      `yaml:"queryPort"`
	RconPort             *int     `yaml:"rconPort"`
	SeamlessIP           string   `yaml:"seamlessIp"`
	AltSaveDirectoryName string   `yaml:"altSaveDirectoryName"`
	Password             string   `yaml:"password"`
	MaxPlayers           int      `yaml:"maxPlayers"`
	ReservedPlayerSlots  int      `yaml:"reservedPlayerSlots"`
	BattleEye            bool     `yaml:"battleEye"`
	PreExecArgs          string   `yaml:"preExecArgs"`
	PostExecArgs         []string `yaml:"postExecArgs"`
}
