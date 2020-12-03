package elasticsearch

import (
	"context"
	"encoding/json"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/epigos/gh-elections-2020-tweets/internal/config"
	"github.com/epigos/gh-elections-2020-tweets/internal/nlp"
	"log"
	"sync"
	"time"
)

type (
	ES struct {
		client *elasticsearch.Client
		wg     *sync.WaitGroup
		index  string
	}
	user struct {
		Name          string `json:"name"`
		ScreenName    string `json:"screen_name"`
		Location      string `json:"location"`
		Description   string `json:"description"`
		FollowerCount int    `json:"followers_count"`
		ID            string `json:"uid"`
		StatusCount   int    `json:"status_count"`
		DateJoined    string `json:"date_joined"`
	}
	userMention struct {
		Name       string `json:"name"`
		ScreenName string `json:"screen_name"`
		ID         string `json:"uid"`
	}
	doc struct {
		User              user          `json:"user"`
		CreatedAt         time.Time     `json:"created_at"`
		Text              string        `json:"text"`
		QuoteCount        int           `json:"quote_count"`
		ReplyCount        int           `json:"reply_count"`
		RetweetCount      int           `json:"retweet_count"`
		FavoriteCount     int           `json:"favorite_count"`
		Sentiment         string        `json:"sentiment"`
		Polarity          float32       `json:"polarity"`
		Hashtag           string        `json:"hashtag"`
		Party             string        `json:"party"`
		CleanText         string        `json:"clean_text"`
		Place             string        `json:"place"`
		Entities          []nlp.Entity  `json:"entities"`
		Source            string        `json:"source"`
		Language          string        `json:"language"`
		PossiblySensitive bool          `json:"possibly_sensitive"`
		UserMentions      []userMention `json:"user_mentions"`
	}
)

func NewES(index string) *ES {
	es, err := elasticsearch.NewClient(config.GetESConfig())

	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()
	log.Println(res)

	wg := &sync.WaitGroup{}
	return &ES{es, wg, index}
}

func (e *ES) Consume(ch chan *twitter.Tweet) {
	for {
		e.wg.Add(1)
		tweet := <-ch
		log.Printf("Received tweet: %vs\n", tweet)

		go e.indexDocument(tweet)
		e.wg.Wait()
	}
}

func (e *ES) UpdateMapping() {

}

func (e *ES) indexDocument(t *twitter.Tweet) {
	defer e.wg.Done()
	// Build the index document.
	doc := newDoc(t)
	// Set up the request object.
	req := esapi.IndexRequest{
		Index:        e.index,
		DocumentType: "tweet",
		DocumentID:   t.IDStr,
		Body:         esutil.NewJSONReader(doc),
		Refresh:      "true",
	}
	// Perform the request with the client.
	res, err := req.Do(context.Background(), e.client)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID=%d", res.Status(), t.ID)
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and indexed document version.
			log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
		}
	}
}

func newDoc(t *twitter.Tweet) *doc {
	var hashtag string
	if len(t.Entities.Hashtags) > 0 {
		hashtag = t.Entities.Hashtags[0].Text
	}
	createdAt, _ := t.CreatedAtTime()

	var place string
	if t.Place != nil {
		place = t.Place.Name
	}
	analysis := nlp.Analyze(t.Text)

	d := &doc{
		User: user{
			ScreenName:    t.User.ScreenName,
			Name:          t.User.Name,
			Description:   t.User.Description,
			Location:      parseLocation(t.User.Location),
			FollowerCount: t.User.FollowersCount,
			ID:            t.User.IDStr,
			StatusCount:   t.User.StatusesCount,
			DateJoined:    t.User.CreatedAt,
		},
		CreatedAt:         createdAt,
		Text:              t.Text,
		QuoteCount:        t.QuoteCount,
		ReplyCount:        t.ReplyCount,
		RetweetCount:      t.RetweetCount,
		FavoriteCount:     t.FavoriteCount,
		Hashtag:           hashtag,
		Party:             getParty(hashtag),
		Place:             place,
		CleanText:         analysis.CleanText,
		Entities:          analysis.Entities,
		Sentiment:         analysis.Sentiment,
		Polarity:          analysis.Polarity,
		Source:            parseSource(t.Source),
		Language:          analysis.Language,
		PossiblySensitive: t.PossiblySensitive,
		UserMentions:      parseUserMentions(t),
	}

	return d
}
