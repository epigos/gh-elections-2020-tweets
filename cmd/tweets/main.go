package main

import (
	"context"
	"flag"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/epigos/gh-elections-2020-tweets/internal/config"
	"github.com/epigos/gh-elections-2020-tweets/internal/elasticsearch"
	"github.com/epigos/gh-elections-2020-tweets/internal/health"
	"github.com/epigos/gh-elections-2020-tweets/internal/nlp"
	"github.com/epigos/gh-elections-2020-tweets/internal/tweets"
	"log"
)

func main() {
	var (
		environment = flag.String("e", "local", "provide run environment local|dev|prod")
	)
	flag.Parse()
	// initialize config
	config.Init(*environment)
	cfg := config.GetConfig()

	log.Printf("Streaming tweets from the following hashtags...\n%v", config.Hashtags)

	ctx := context.Background()
	nlp.Init(ctx)

	es := elasticsearch.NewES("tweets")
	stream := tweets.NewStreamListener(cfg)

	ch := make(chan *twitter.Tweet)

	go health.Start("")

	go es.Consume(ch)

	log.Fatal(stream.Listen(ch))
}
