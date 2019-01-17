package config

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

func LoadOrInit(file string) (*Config, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			if content, err := yaml.Marshal(DefaultConfig()); err != nil {
				return nil, err
			} else if err := ioutil.WriteFile(file, content, 0644); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	conf := &Config{}
	err = yaml.Unmarshal(content, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func DefaultConfig() *Config {
	return &Config{
		Port:    8080,
		Servers: make([]ServerOptions, 0),
	}
}

type Config struct {
	Port    int             `yaml:"port"`
	Servers []ServerOptions `yaml:"servers"`
}

type ServerOptions struct {
	Pid          int             `yaml:"pid"`
	System       SystemOptions   `yaml:"system"`
	Gameplay     GameplayOptions `yaml:"gameplay"`
	PreExecArgs  string          `yaml:"preExecArgs"`
	PostExecArgs string          `yaml:"postExecArgs"`
	ExecString   string          `yaml:"execString"`
}

type SystemOptions struct {
	ServerX              int    `yaml:"serverX"`
	ServerY              int    `yaml:"serverY"`
	Port                 int    `yaml:"port"`
	QueryPort            int    `yaml:"queryPort"`
	AltSaveDirectoryName string `yaml:"altSaveDirectoryName"`
	MaxPlayers           int    `yaml:"maxPlayers"`
	ReservedPlayerSlots  int    `yaml:"reservedPlayerSlots"`
	SeamlessIP           string `yaml:"seamlessIp"`
	RconPort             *int   `yaml:"rconPort"`
	IPAddress            string `yaml:"IpAddress"`
	HasPassword          bool   `yaml:"hasPassword"`
	BattleEye            bool   `yaml:"battleEye"`
}

type GameplayOptions struct {
	Pve                          bool `yaml:"pve"`
	AllowAnyoneBabyImprintCuddle bool `yaml:"allowAnyoneBabyImprintCuddle"`
	EnablePvpGamma               bool `yaml:"enablePvpGamma"`
	ShowFloatingDamageText       bool `yaml:"showFloatingDamageText"`
	EnableCrosshair              bool `yaml:"enableCrosshair"`
	AllowThirdPersonPlayer       bool `yaml:"allowThirdPersonPlayer"`
	MapPlayerLocation            bool `yaml:"mapPlayerLocation"`
}
