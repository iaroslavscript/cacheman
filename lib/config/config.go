package config

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

//type config struct { // TODO
type Config struct {
	BindAddr                   string `json:"bind_addr"`
	ExpiresDefaultDurationSec  int64  `json:"expires_default_duration_sec"`
	ReplicationRotateEveryMs   int64  `json:"replication_rotate_every_ms"`
	ShedulerDelExpiredEverySec int64  `json:"sheduler_del_expired_every_sec"`
	ShedulerExpiredQuequeSize  int64  `json:"sheduler_expired_queque_size"`
}

var instance *Config
var once sync.Once
var m sync.Mutex

func GetConfig() *Config {

	once.Do(func() {
		instance = defaultConfig()
	})

	return instance
}

func LoadConfig(filepath string) error {

	var cfg Config
	var err error

	m.Lock()
	defer m.Unlock()

	cfg, err = loadConfig(filepath)
	*instance = cfg

	return err
}

func Validate() error {
	// no-op at current version
	// TODO add validation logic
	return nil
}

func defaultConfig() *Config {

	return &Config{
		BindAddr:                   "0.0.0.0:8080",
		ExpiresDefaultDurationSec:  30 * 60,
		ReplicationRotateEveryMs:   1000,
		ShedulerDelExpiredEverySec: 60,
		ShedulerExpiredQuequeSize:  1000,
	}
}

func loadConfig(filepath string) (Config, error) {

	var cfg Config
	var data []byte
	var err error

	if data, err = ioutil.ReadFile(filepath); err != nil {
		return cfg, err
	}

	// unmarshall it
	if err = json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, err
}
