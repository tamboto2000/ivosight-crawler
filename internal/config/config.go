package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type MongoDB struct {
	Host     string
	Port     string
	Username string
	Password string
}

func (mongo MongoDB) ToURL() string {
	url := fmt.Sprintf("mongodb://%s:%s@%s:%s", mongo.Username, mongo.Password, mongo.Host, mongo.Port)
	return url
}

type Crawler struct {
	MaxThreadCount         int
	UseProxy               bool
	UseProxyscrape         bool
	UseProxyList           bool
	ProxyList              []string
	RandomRunIntervalRange []int64
	RespectRobotsTxt       bool
}

type Config struct {
	MongoDB MongoDB
	Crawler Crawler
}

const (
	defaultMaxThreadCount = 12
	defaultUseProxy       = false
	defaultUseProxyScrape = false
)

var defaultRandomRunIntervalRange = []int64{60, 300}

func LoadConfig() (Config, error) {
	return parseConfig()
}

func parseConfig() (Config, error) {
	var cfg Config
	if err := godotenv.Load(); err != nil {
		return cfg, err
	}

	mongoHost := os.Getenv("MONGO_HOST")
	mongoPort := os.Getenv("MONGO_PORT")
	mongoUser := os.Getenv("MONGO_USERNAME")
	mongoPwd := os.Getenv("MONGO_PASSWORD")

	maxThreadCount := os.Getenv("MAX_THREAD_COUNT")
	useProxy := os.Getenv("USE_PROXY")
	userProxyScrape := os.Getenv("USE_PROXYSCRAPE")
	useProxyList := os.Getenv("USE_PROXY_LIST")
	proxyList := os.Getenv("PROXY_LIST")
	randomRunInterval := os.Getenv("RANDOM_RUN_INTERVAL_RANGE")

	mongoCfg := MongoDB{
		Host:     mongoHost,
		Port:     mongoPort,
		Username: mongoUser,
		Password: mongoPwd,
	}

	crawlerCfg := Crawler{
		MaxThreadCount:         strToInt(maxThreadCount, defaultMaxThreadCount),
		UseProxy:               strToBool(useProxy, false),
		UseProxyscrape:         strToBool(userProxyScrape, false),
		UseProxyList:           strToBool(useProxyList, false),
		ProxyList:              strToStrSlice(proxyList, ",", []string{}),
		RandomRunIntervalRange: strToInt64Slice(randomRunInterval, "-", defaultRandomRunIntervalRange),
	}

	cfg.MongoDB = mongoCfg
	cfg.Crawler = crawlerCfg

	return cfg, nil
}

func strToInt(str string, def int) int {
	i, err := strconv.Atoi(str)
	if err != nil || i <= 0 {
		return def
	}

	return i
}

func strToBool(str string, def bool) bool {
	str = strings.ToLower(str)
	switch str {
	case "true", "1":
		return true

	case "false", "0":
		return false

	default:
		return def
	}
}

func strToStrSlice(str string, delim string, def []string) []string {
	split := strings.Split(str, delim)
	if len(split) == 0 {
		return def
	}

	return split
}

func strToInt64Slice(str string, delim string, def []int64) []int64 {
	var nums []int64
	split := strings.Split(str, delim)
	for _, str := range split {
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return def
		}

		nums = append(nums, i)
	}

	return nums
}
