package conf

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Handler struct {
		MaxResponseTime     int   `yaml:"max_response_time"`
		MaxSiteResponseTime int   `yaml:"max_site_response_time"`
		MaxUrl              int   `yaml:"max_url"`
		MaxError            int   `yaml:"max_error"`
		NumGoroutineList    []int `yaml:"num_goroutine_list"`
	} `yaml:"handler"`
}

func ReadConfig(cfg *Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}
