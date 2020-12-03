package config

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/spf13/viper"
	"log"
	"strings"
)

var (
	config   *viper.Viper
	NDCTags  []string
	NPPTags  []string
	Hashtags []string
)

// Init is an exported method that takes the environment starts the viper
// (external lib) and returns the configuration struct.
func Init(env string) {
	var err error
	config = viper.New()
	config.SetConfigType("yml")
	config.SetConfigName(env)
	config.AddConfigPath("config/")
	config.AutomaticEnv()

	// set default configs
	config.SetDefault("env", env)
	config.SetDefault("ELASTICSEARCH_HOST", "http://localhost:9200")
	config.SetDefault("DEFAULT_HASHTAGS", []string{})

	// read config file
	err = config.ReadInConfig()
	if err != nil {
		log.Printf("Failed to read the configuration file: %s.yml\n", env)
	}
	NDCTags = getNDCTags()
	NPPTags = getNPPTags()
	Hashtags = getHashtags()
}

// GetConfig return config object
func GetConfig() *viper.Viper {
	return config
}

func getNDCTags() []string {
	tagStr := config.GetString("NDC_HASHTAGS")

	return strings.Split(tagStr, ",")
}

func getNPPTags() []string {
	tagStr := config.GetString("NPP_HASHTAGS")

	return strings.Split(tagStr, ",")
}

func getHashtags() []string {
	tagStr := config.GetString("DEFAULT_HASHTAGS")

	defaultTags := strings.Split(tagStr, ",")

	tags := append(defaultTags, NPPTags...)
	tags = append(tags, NDCTags...)

	return tags
}

func GetESConfig() elasticsearch.Config {
	esAddrStr := config.GetString("ELASTICSEARCH_HOST")
	cfg := elasticsearch.Config{
		Addresses: strings.Split(esAddrStr, ","),
		Username:  config.GetString("ELASTICSEARCH_USER"),
		Password:  config.GetString("ELASTICSEARCH_PWD"),
	}

	return cfg
}
