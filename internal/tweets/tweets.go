package tweets

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/epigos/gh-elections-2020-tweets/internal/config"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type StreamListener struct {
	client *twitter.Client
	config *viper.Viper
}

func NewStreamListener(cfg *viper.Viper) *StreamListener {
	oauthConfig := oauth1.NewConfig(
		cfg.GetString("TWITTER_CONSUMER_KEY"),
		cfg.GetString("TWITTER_CONSUMER_SECRET"))

	token := oauth1.NewToken(
		cfg.GetString("TWITTER_ACCESS_TOKEN"),
		cfg.GetString("TWITTER_ACCESS_TOKEN_SECRET"))
	httpClient := oauthConfig.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	st := &StreamListener{client, cfg}

	return st
}

func (s *StreamListener) Listen(msgCh chan *twitter.Tweet) error {
	log.Println("Starting streaming...")

	params := &twitter.StreamFilterParams{
		Track:         config.Hashtags,
		StallWarnings: twitter.Bool(true),
	}

	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		if !strings.HasPrefix(tweet.Text, "RT @") {
			go func() { msgCh <- tweet }()
		}
	}
	demux.Event = func(event *twitter.Event) {
		log.Printf("%#v\n", event)
	}

	stream, err := s.client.Streams.Filter(params)
	if err != nil {
		return err
	}

	go demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	log.Println("Stopping Stream...")
	stream.Stop()
	return nil
}
