package nlp

import (
	language "cloud.google.com/go/language/apiv1"
	"context"
	"github.com/bbalet/stopwords"
	langpb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"log"
	"regexp"
	"strings"
)

var (
	client *language.Client
	ctx    context.Context
	rmLink = regexp.MustCompile(`https*\S+`)
	rmAt   = regexp.MustCompile(`@\S+`)
	rmHash = regexp.MustCompile(`#\S+`)
	rmW    = regexp.MustCompile(`\'\w+`)
	rmD    = regexp.MustCompile(`\w*\d+\w*`)
	rmS    = regexp.MustCompile(`\s{2,}`)
)

// NLPResult represents processed Message and it's content
type (
	Result struct {
		Sentiment string `json:"score"`
		Polarity  float32
		Entities  []Entity `json:"entities"`
		CleanText string   `json:"clean_text"`
		Language  string   `json:"language"`
	}
	Entity struct {
		Text  string `json:"text"`
		Label string `json:"label"`
	}
)

func Init(cx context.Context) {
	ctx = cx
	// Creates a client.
	var err error
	client, err = language.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
}

func Analyze(t string) *Result {
	t = cleanText(t)

	annotate, err := client.AnnotateText(ctx, &langpb.AnnotateTextRequest{
		Document: &langpb.Document{
			Source: &langpb.Document_Content{
				Content: t,
			},
			Type: langpb.Document_PLAIN_TEXT,
		},
		EncodingType: langpb.EncodingType_UTF8,
		Features: &langpb.AnnotateTextRequest_Features{
			ExtractSyntax:            false,
			ExtractEntities:          true,
			ExtractDocumentSentiment: true,
			ExtractEntitySentiment:   false,
			ClassifyText:             false,
		},
	})

	if err != nil {
		log.Printf("Failed to annotate text: %v\n", err)
		return &Result{}
	}

	r := &Result{
		Sentiment: getSentiment(annotate.DocumentSentiment),
		Polarity:  annotate.DocumentSentiment.Score,
		Entities:  getEntities(annotate.Entities),
		CleanText: t,
		Language:  annotate.Language,
	}
	return r
}

func getSentiment(s *langpb.Sentiment) string {

	if s.Score < 0 {
		return "negative"
	} else if s.Score == 0 {
		return "neutral"
	}
	return "positive"
}

func getEntities(results []*langpb.Entity) []Entity {

	var entities []Entity
	for _, e := range results {
		entities = append(entities, Entity{
			Text:  e.Name,
			Label: e.Type.String(),
		})
	}
	return entities
}

func cleanText(t string) string {
	x := strings.ToLower(t)

	x = rmLink.ReplaceAllString(x, " ")
	//x = rmAt.ReplaceAllString(x, " ")
	x = rmHash.ReplaceAllString(x, " ")
	x = rmW.ReplaceAllString(x, "")
	x = rmD.ReplaceAllString(x, "")
	x = rmS.ReplaceAllString(x, " ")

	x = stopwords.CleanString(x, "en", true)
	return x
}
