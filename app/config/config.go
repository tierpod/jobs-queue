// Package config contains configuration description.
package config

import (
	"io/ioutil"
	"regexp"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// Config contains all parameters.
type Config struct {
	Workers         int           `yaml:"workers"`
	LogDebug        bool          `yaml:"log_debug"`
	LogDatetime     bool          `yaml:"log_datetime"`
	Socket          string        `yaml:"socket"`
	QueueSize       int           `yaml:"queue_size"`
	CacheExpire     time.Duration `yaml:"cache_expire"`
	CacheDeleteMode string        `yaml:"cache_delete_mode"`
	CacheExludes    []string      `yaml:"cache_excludes"`
	Jobs            []string      `yaml:"jobs"`
	CacheExcludesRe []*regexp.Regexp
}

// Load loads file `path` and returns configuration. Converts CacheExpire to seconds.
func Load(path string) (*Config, error) {
	var c Config

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	// convert int to second
	c.CacheExpire *= time.Second

	if c.CacheDeleteMode == "" {
		c.CacheDeleteMode = "expire"
	}

	// compile regexp
	for _, r := range c.CacheExludes {
		c.CacheExcludesRe = append(c.CacheExcludesRe, regexp.MustCompile(r))
	}

	return &c, nil
}
