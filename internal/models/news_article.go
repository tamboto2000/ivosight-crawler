package models

import (
	"encoding/json"
	"time"
)

type ArticleSource string

const (
	Liputan6 = "Liputan6.com"
	Detik    = "Detik.com"
)

type ArticleAuthor struct {
	Name       string `bson:"name" json:"name"`
	ProfileURL string `bson:"profile_url" json:"profile_url"`
}

type ArticleImageContent struct {
	URL     string `bson:"url" json:"url"`
	Title   string `bson:"title" json:"title"`
	Caption string `bson:"caption" json:"caption"`
	Alt     string `bson:"alt" json:"alt"`
	Width   int    `bson:"width" json:"width"`
	Height  int    `bson:"height" json:"height"`
}

type ArticleVideoContent struct {
	URL         string `bson:"url" json:"url"`
	EmbeddedURL string `bson:"embedded_url" json:"embedded_url"`
	Title       string `bson:"title" json:"title"`
	Description string `bson:"description" json:"description"`
	Duration    int64  `bson:"duration" json:"duration"`
	Thumbnail   string `bson:"thumbnail" json:"thumbnail"`
}

type ArticleContent struct {
	Type string          `bson:"type" json:"type"`
	Data json.RawMessage `bson:"data" json:"data"`
}

type RelatedArticle struct {
	Title       string              `bson:"title" json:"title"`
	ArticleLink string              `bson:"related_article"`
	Thumbnail   ArticleImageContent `bson:"thumbnail" json:"thumbnail"`
}

type NewsArticle struct {
	ID              string           `bson:"_id" json:"-"`
	Source          ArticleSource    `bson:"source" json:"source"`
	Headline        string           `bson:"headline" json:"headline"`
	Description     string           `bson:"description" json:"description"`
	PublishedAt     time.Time        `bson:"published_at" json:"published_at"`
	UpdatedAt       time.Time        `bson:"updated_at" json:"updated_at"`
	Author          ArticleAuthor    `bson:"article_author" json:"article_author"`
	Contents        []ArticleContent `bson:"contents" json:"contents"`
	RelatedArticles []RelatedArticle `bson:"related_articles" json:"related_articles"`
}
