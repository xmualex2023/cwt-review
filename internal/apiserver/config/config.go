package config

import (
	"os"
	"reflect"
	"time"

	"gopkg.in/yaml.v3"
)

// TODO: 后面可以同步Viper来进行配置管理
type Config struct {
	Server struct {
		Mode string `yaml:"mode"`
		HTTP struct {
			Address string        `yaml:"address"`
			Timeout time.Duration `yaml:"timeout"`
		} `yaml:"http"`
	} `yaml:"server"`

	Pprof struct {
		Address string `yaml:"address"`
	} `yaml:"pprof"`

	MongoDB struct {
		URI      string `yaml:"uri"`
		Database string `yaml:"database"`
	} `yaml:"mongodb"`

	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`

	JWT struct {
		Secret string        `yaml:"secret"`
		Expire time.Duration `yaml:"expire"`
	} `yaml:"jwt"`

	LLM struct {
		APIKey   string `yaml:"api_key"`
		Endpoint string `yaml:"endpoint"`
	} `yaml:"llm"`

	Metrics struct {
		PullHost        string    `yaml:"pull_host"`
		PushIntervalSec int       `yaml:"push_interval_sec"`
		URL             string    `yaml:"url"`
		Instance        string    `yaml:"instance"`
		Job             string    `yaml:"job"`
		Buckets         []float64 `yaml:"buckets"`
	} `yaml:"metrics"`

	RateLimit struct {
		MaxRequests int64         `yaml:"max_requests"`
		Duration    time.Duration `yaml:"duration"`
	} `yaml:"rate_limit"`

	Worker struct {
		Count int `yaml:"count"`
	} `yaml:"worker"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: struct {
			Mode string `yaml:"mode"`
			HTTP struct {
				Address string        `yaml:"address"`
				Timeout time.Duration `yaml:"timeout"`
			} `yaml:"http"`
		}{
			Mode: "debug",
			HTTP: struct {
				Address string        `yaml:"address"`
				Timeout time.Duration `yaml:"timeout"`
			}{
				Address: ":8080",
				Timeout: 30 * time.Second,
			},
		},
		Pprof: struct {
			Address string `yaml:"address"`
		}{
			Address: ":6060",
		},
		MongoDB: struct {
			URI      string `yaml:"uri"`
			Database string `yaml:"database"`
		}{
			URI:      "mongodb://localhost:27017",
			Database: "myapp",
		},
		Redis: struct {
			Addr     string `yaml:"addr"`
			Password string `yaml:"password"`
			DB       int    `yaml:"db"`
		}{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		JWT: struct {
			Secret string        `yaml:"secret"`
			Expire time.Duration `yaml:"expire"`
		}{
			Secret: "default-secret-key",
			Expire: 24 * time.Hour,
		},
		LLM: struct {
			APIKey   string `yaml:"api_key"`
			Endpoint string `yaml:"endpoint"`
		}{
			APIKey:   "",
			Endpoint: "https://api.openai.com/v1",
		},
		Metrics: struct {
			PullHost        string    `yaml:"pull_host"`
			PushIntervalSec int       `yaml:"push_interval_sec"`
			URL             string    `yaml:"url"`
			Instance        string    `yaml:"instance"`
			Job             string    `yaml:"job"`
			Buckets         []float64 `yaml:"buckets"`
		}{
			PullHost:        ":9090",
			PushIntervalSec: 15,
			URL:             "http://localhost:9091",
			Instance:        "default",
			Job:             "i18n-apiserver",
			Buckets:         []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		RateLimit: struct {
			MaxRequests int64         `yaml:"max_requests"`
			Duration    time.Duration `yaml:"duration"`
		}{
			MaxRequests: 100,
			Duration:    time.Minute,
		},
		Worker: struct {
			Count int `yaml:"count"`
		}{
			Count: 5,
		},
	}
}

// MergeConfig 使用反射合并配置
func mergeConfig(dst, src interface{}) {
	dstValue := reflect.ValueOf(dst).Elem()
	srcValue := reflect.ValueOf(src).Elem()

	for i := 0; i < dstValue.NumField(); i++ {
		dstField := dstValue.Field(i)
		srcField := srcValue.Field(i)

		switch dstField.Kind() {
		case reflect.Struct:
			if dstField.CanAddr() && srcField.CanAddr() {
				mergeConfig(dstField.Addr().Interface(), srcField.Addr().Interface())
			}
		default:
			if !srcField.IsZero() {
				dstField.Set(srcField)
			}
		}
	}
}

func Load(path string) (*Config, error) {
	// 首先加载默认配置
	cfg := DefaultConfig()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	tmpCfg := &Config{}
	if err := yaml.Unmarshal(data, tmpCfg); err != nil {
		return nil, err
	}

	mergeConfig(cfg, tmpCfg)
	return cfg, nil
}
