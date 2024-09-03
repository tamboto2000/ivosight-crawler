package detik

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tamboto2000/ivosight-crawler/pkg/htmlutil"
	"golang.org/x/net/html"
)

type ArticleType string

const (
	SinglePageArticle    ArticleType = "singlepage"
	MultiplePhotoArticle ArticleType = "multiplefoto"
	VideoArticle         ArticleType = "video"
)

type articleJsonScript struct {
	Headline string `json:"headline"`
	Image    struct {
		URL             string `json:"url"`
		ContentLocation string `json:"contentLocation"`
	} `json:"image"`
	Video struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		ThumbnailURL string `json:"thumbnailUrl"`
		Duration     string `json:"duration"`
		Keywords     string `json:"keywords"`
		URL          string `json:"url"`
		ContentURL   string `json:"contentUrl"`
		EmbedURL     string `json:"embedUrl"`
	} `json:"video"`
	DatePublished time.Time `json:"datePublished"`
	DateModified  time.Time `json:"dateModified"`
	Author        struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"author"`
	Description string `json:"description"`
}

type ArticleContentType string

const (
	SectionTitle          ArticleContentType = "section-title"
	ParagraphText         ArticleContentType = "paragraph-text"
	Image                 ArticleContentType = "image"
	Video                 ArticleContentType = "video"
	ReferencedArticleLink ArticleContentType = "referenced-article-link"
	PublishedFrom         ArticleContentType = "published-from"
)

type ContentImage struct {
	URL     string
	Alt     string
	Title   string
	Caption string
}

type ContentVideo struct {
	URL          string
	EmbeddedURL  string
	Title        string
	Description  string
	Duration     int64
	ThumbnailURL string
}

type Article struct {
	Type            ArticleType
	Link            string
	Headline        string
	HeadlineImage   *ContentImage
	Description     string
	Author          string
	PublishedFrom   string
	PublishedAt     time.Time
	UpdatedAt       time.Time
	Contents        []ArticleContent
	RelatedArticles []RelatedArticle
}

type ArticleContent struct {
	Type ArticleContentType
	Data any
}

func (art ArticleContent) String() string {
	str, ok := art.Data.(string)
	if !ok {
		panic("not a string")
	}

	return str
}

func (art ArticleContent) ContentImage() ContentImage {
	img, ok := art.Data.(ContentImage)
	if !ok {
		panic("not an image")
	}

	return img
}

func (art ArticleContent) ReferencedArticle() ReferencedArticle {
	ref, ok := art.Data.(ReferencedArticle)
	if !ok {
		panic("not a referenced article")
	}

	return ref
}

type ReferencedArticle struct {
	Headline    string
	ArticleLink string
}

type RelatedArticle struct {
	Title       string
	ArticleLink string
}

type newsArticleParser struct {
	art                Article
	isJsonScriptParsed bool
	contVid            ContentVideo
}

func (nap *newsArticleParser) parseArticle(ctx context.Context, dtk *Detik, link string) (Article, error) {
	node, err := dtk.commonReq(ctx, link)
	if err != nil {
		return Article{}, err
	}

	nap.art.Link = link

	htmlutil.WalkSkipNodes(node, nap.walkNodesNewsArticle)
	return nap.art, nil
}

func (nap *newsArticleParser) walkNodesNewsArticle(node *html.Node) (bool, bool) {
	if nap.isArticleTypeMeta(node) {
		nap.parseArticleType(node)
		return false, true
	}

	if nap.isAuthor(node) {
		nap.parseAuthor(node)
		return false, true
	}

	if nap.art.Type == VideoArticle {
		if nap.isNewsVideoDuration(node) {
			nap.parseNewsVideoDuration(node)
			return false, true
		}
	}

	if nap.isJsonScript(node) {
		nap.parseJsonScript(node)
		return false, true
	}

	switch nap.art.Type {
	case SinglePageArticle:
		if nap.isMainContent(node) {
			nap.parseMainContent(node)
			return false, true
		}

		if nap.isHeadlineImageFig(node) {
			nap.parseHeadlineImage(node)
			return false, true
		}

	case MultiplePhotoArticle:
		if nap.isNewsFotoMainContent(node) {
			nap.parseNewsFotoMainContent(node)
			return false, true
		}

	case VideoArticle:
		if nap.isNewsVideoMainContent(node) {
			nap.parseNewsVideoMainContent(node)
			return false, true
		}
	}

	if nap.isRelatedArticleList(node) {
		nap.parseRelatedArticleList(node)
		return false, true
	}

	return true, true
}

func (nap *newsArticleParser) isMainContent(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "detail__body-text itp_bodycontent" {
				return true
			}
		}
	}

	return false
}

func (nap *newsArticleParser) parseMainContent(node *html.Node) {
	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if nap.isStrongTag(node) {
			if nap.art.PublishedFrom != "" {
				return false, true
			}

			htmlutil.WalkNodes(node, func(node *html.Node) bool {
				if node.Type == html.TextNode {
					nap.art.PublishedFrom = node.Data
				}

				return true
			})

			return false, true
		}

		if nap.isNodeParagraph(node) {
			nap.parseParagraph(node)
			return false, true
		}

		if nap.isImage(node) {
			nap.parseImage(node)
			return false, true
		}

		if nap.isSectionTitle(node) {
			nap.parseSectionTitle(node)
			return false, true
		}

		if nap.isNodeReferencedArticle(node) {
			nap.parseReferencedArticle(node)
			return false, true
		}

		return true, true
	})
}

func (nap *newsArticleParser) isHeadlineImageFig(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "figure" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "detail__media-image" {
				return true
			}
		}
	}

	return false
}

func (nap *newsArticleParser) parseHeadlineImage(node *html.Node) {
	nap.art.HeadlineImage = &ContentImage{}
	htmlutil.WalkNodes(node, func(node *html.Node) bool {
		if node.Type == html.ElementNode {
			if node.Data == "img" {
				for _, attr := range node.Attr {
					switch attr.Key {
					case "src":
						nap.art.HeadlineImage.URL = attr.Val

					case "alt":
						nap.art.HeadlineImage.Alt = attr.Val

					case "title":
						nap.art.HeadlineImage.Title = attr.Val
					}
				}
			}

			if node.Data == "figcaption" {
				htmlutil.WalkNodes(node, func(node *html.Node) bool {
					if node.Type == html.TextNode {
						nap.art.HeadlineImage.Caption += node.Data
					}

					return true
				})

				nap.art.HeadlineImage.Caption = strings.TrimSpace(nap.art.HeadlineImage.Caption)
			}
		}

		return true
	})
}

func (nap *newsArticleParser) isNodeParagraph(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "p" && len(node.Attr) == 0 {
		return true
	}

	return false
}

func (nap *newsArticleParser) parseParagraph(node *html.Node) {
	artType := ParagraphText
	var paragraph string
	var data any

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if node.Type == html.TextNode {
			paragraph += node.Data
			return false, true
		}

		if node.Type == html.ElementNode {
			if node.Data == "p" {
				return true, true
			}

			if node.Data == "a" {
				for _, attr := range node.Attr {
					if attr.Key == "href" {
						data = ContentVideo{
							EmbeddedURL: attr.Val,
						}
					}

					// embedded video
					if attr.Key == "class" && attr.Val == "embed video20detik" {
						artType = Video
						break
					}
				}

				if artType != ParagraphText {
					return false, false
				}
			}

			var buff bytes.Buffer
			html.Render(&buff, node)

			paragraph += buff.String()

			return false, true
		}

		return true, true
	})

	if paragraph != "" && artType == ParagraphText {
		nap.art.Contents = append(nap.art.Contents, ArticleContent{
			Type: ParagraphText,
			Data: paragraph,
		})
	}

	if artType == Video {
		nap.art.Contents = append(nap.art.Contents, ArticleContent{
			Type: Video,
			Data: data,
		})
	}
}

func (nap *newsArticleParser) isSectionTitle(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "h2" {
		return true
	}

	return false
}

func (nap *newsArticleParser) parseSectionTitle(node *html.Node) {
	var sectionTitle string

	htmlutil.WalkNodes(node, func(node *html.Node) bool {
		if node.Type == html.ElementNode {
			if node.Data == "h2" {
				return true
			}

			var buff bytes.Buffer
			html.Render(&buff, node)
			sectionTitle += buff.String()
		}

		if node.Type == html.TextNode {
			sectionTitle += node.Data
		}

		return true
	})

	nap.art.Contents = append(nap.art.Contents, ArticleContent{
		Type: SectionTitle,
		Data: sectionTitle,
	})
}

func (nap *newsArticleParser) isNodeReferencedArticle(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "lihatjg" {
				return true
			}
		}
	}

	return false
}

func (nap *newsArticleParser) parseReferencedArticle(node *html.Node) {
	var refart ReferencedArticle
	htmlutil.WalkNodes(node, func(node *html.Node) bool {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					refart.ArticleLink = attr.Val
				}
			}

			htmlutil.WalkNodes(node, func(node *html.Node) bool {
				if node.Type == html.TextNode {
					refart.Headline = node.Data
				}

				return true
			})

			return false
		}

		return true
	})

	nap.art.Contents = append(nap.art.Contents, ArticleContent{
		Type: ReferencedArticleLink,
		Data: refart,
	})
}

func (nap *newsArticleParser) isStrongTag(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "strong" {
		return true
	}

	return false
}

func (nap *newsArticleParser) isJsonScript(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "script" {
		for _, attr := range node.Attr {
			if attr.Key == "type" && attr.Val == "application/ld+json" {
				return true
			}
		}
	}

	return false
}

var newsartJsonRgx = regexp.MustCompile(`"@type"\s?:\s?"NewsArticle"`)

func (nap *newsArticleParser) parseJsonScript(node *html.Node) {
	if nap.isJsonScriptParsed {
		return
	}

	var script string
	isTypeArticle := false

	htmlutil.WalkNodes(node, func(node *html.Node) bool {
		if node.Type == html.TextNode {
			script += node.Data
			if newsartJsonRgx.MatchString(node.Data) {
				isTypeArticle = true
			}
		}

		return true
	})

	if isTypeArticle {
		var obj articleJsonScript
		json.Unmarshal([]byte(script), &obj)

		nap.art.Headline = obj.Headline
		nap.art.Description = obj.Description
		nap.art.PublishedAt = obj.DatePublished
		nap.art.UpdatedAt = obj.DateModified

		if nap.art.Type == MultiplePhotoArticle {
			nap.art.PublishedFrom = obj.Image.ContentLocation
		}

		if nap.art.Type == VideoArticle {
			vid := obj.Video
			nap.contVid = ContentVideo{
				Title:        vid.Name,
				Description:  vid.Description,
				ThumbnailURL: vid.ThumbnailURL,
				URL:          vid.ContentURL,
				EmbeddedURL:  vid.EmbedURL,
				Duration:     nap.contVid.Duration,
			}
		}

		nap.isJsonScriptParsed = true
	}
}

func (nap *newsArticleParser) isRelatedArticleList(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "id" && attr.Val == "bt_tkt" {
				return true
			}
		}
	}

	return false
}

func (nap *newsArticleParser) parseRelatedArticleList(node *html.Node) {
	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if node.Type == html.ElementNode && node.Data == "article" {
			htmlutil.WalkNodes(node, func(node *html.Node) bool {
				if node.Type == html.ElementNode && node.Data == "a" {
					var link, title string
					for _, attr := range node.Attr {
						switch attr.Key {
						case "dtr-ttl":
							title = attr.Val

						case "href":
							link = attr.Val
						}
					}

					nap.art.RelatedArticles = append(nap.art.RelatedArticles, RelatedArticle{
						Title:       title,
						ArticleLink: link,
					})
				}

				return true
			})

			return false, true
		}

		return true, true
	})
}

func (nap *newsArticleParser) isImage(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "pic_artikel_sisip" {
				return true
			}
		}
	}

	return false
}

func (nap *newsArticleParser) parseImage(node *html.Node) {
	var contentImg ContentImage

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if node.Type == html.ElementNode {
			if node.Data == "img" {
				for _, attr := range node.Attr {
					switch attr.Key {
					case "src":
						contentImg.URL = attr.Val

					case "alt":
						contentImg.Alt = attr.Val

					case "title":
						contentImg.Title = attr.Val
					}
				}
			}

			if node.Data == "span" {
				var caption string
				htmlutil.WalkNodes(node, func(node *html.Node) bool {
					if node.Type == html.TextNode {
						caption += node.Data
					}

					return true
				})

				contentImg.Caption = caption

				return false, false
			}
		}

		return true, true
	})

	nap.art.Contents = append(nap.art.Contents, ArticleContent{
		Type: Image,
		Data: contentImg,
	})
}

func (nap *newsArticleParser) isNewsFotoMainContent(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "article" {
		return false
	}

	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "detail" {
			return true
		}
	}

	return false
}

func (nap *newsArticleParser) parseNewsFotoMainContent(node *html.Node) {
	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if nap.isNewsFotoDetailHeader(node) {
			nap.parseNewsFotoDetailHeader(node)
			return false, true
		}

		if nap.isNodeParagraph(node) {
			nap.parseParagraph(node)
			return false, true
		}

		if nap.isNewsFotoImageContent(node) {
			nap.parseNewsFotoImageContent(node)
			return false, true
		}

		return true, true
	})
}

func (nap *newsArticleParser) isNewsFotoDetailHeader(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "div" {
		return false
	}

	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "detail__header" {
			return true
		}
	}

	return false
}

func (nap *newsArticleParser) parseNewsFotoDetailHeader(node *html.Node) {
	htmlutil.WalkNodes(node, func(node *html.Node) bool {
		if node.Type == html.ElementNode && node.Data == "h1" {
			var title string

			htmlutil.WalkNodes(node, func(node *html.Node) bool {
				if node.Type == html.TextNode {
					title += node.Data
				}

				return true
			})

			title = strings.TrimSpace(title)
			nap.art.Headline = title

			return false
		}

		return true
	})
}

func (nap *newsArticleParser) isNewsFotoImageContent(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "div" {
		return false
	}

	for _, attr := range node.Attr {
		if attr.Key == "id" && attr.Val == "slider-foto__detail" {
			return true
		}
	}

	return false
}

func (nap *newsArticleParser) parseNewsFotoImageContent(node *html.Node) {
	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if node.Type == html.ElementNode && node.Data == "figure" {
			var contentImg ContentImage

			htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
				if node.Type != html.ElementNode {
					return true, true
				}

				switch node.Data {
				case "img":
					for _, attr := range node.Attr {
						switch attr.Key {
						case "alt":
							contentImg.Alt = attr.Val

						case "title":
							contentImg.Title = attr.Val

						case "src":
							contentImg.URL = attr.Val

						case "data-lazy":
							if attr.Val != "" {
								contentImg.URL = attr.Val
							}
						}
					}

					return false, true

				case "figcaption":
					var caption string
					htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
						if node.Data == "figcaption" {
							return true, true
						}

						if node.Data == "div" {
							return true, true
						}

						if node.Type == html.TextNode {
							caption += node.Data
						}

						if node.Type == html.ElementNode {
							var buff bytes.Buffer
							html.Render(&buff, node)

							caption += buff.String()

							return false, true
						}

						return true, true
					})

					caption = strings.TrimSpace(caption)
					contentImg.Caption = caption

					return false, true
				}

				return true, true
			})

			nap.art.Contents = append(nap.art.Contents, ArticleContent{
				Type: Image,
				Data: contentImg,
			})

			return false, true
		}

		return true, true
	})
}

func (nap *newsArticleParser) isAuthor(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "meta" {
		for _, attr := range node.Attr {
			if attr.Key == "name" && attr.Val == "author" {
				return true
			}
		}
	}

	return false
}

func (nap *newsArticleParser) parseAuthor(node *html.Node) {
	for _, attr := range node.Attr {
		if attr.Key == "content" {
			nap.art.Author = attr.Val
			return
		}
	}
}

func (nap *newsArticleParser) isNewsVideoDuration(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "meta" {
		for _, attr := range node.Attr {
			if attr.Key == "name" && attr.Val == "duration" {
				return true
			}
		}
	}

	return false
}

func (nap *newsArticleParser) parseNewsVideoDuration(node *html.Node) {
	for _, attr := range node.Attr {
		if attr.Key == "content" {
			d, _ := strconv.ParseInt(attr.Val, 10, 64)
			nap.contVid.Duration = d
		}
	}
}

func (nap *newsArticleParser) isNewsVideoMainContent(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "detail" {
				return true
			}
		}
	}

	return false
}

func (nap *newsArticleParser) parseNewsVideoMainContent(node *html.Node) {
	// By now, the video is already parsed, so we will append it to
	// nap.art.Contents
	nap.art.Contents = append(nap.art.Contents, ArticleContent{
		Type: Video,
		Data: nap.contVid,
	})

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if nap.isNodeParagraph(node) {
			nap.parseParagraph(node)
			return false, true
		}

		return true, true
	})
}

func (nap *newsArticleParser) isArticleTypeMeta(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "meta" {
		for _, attr := range node.Attr {
			if attr.Key == "name" && attr.Val == "articletype" {
				return true
			}
		}
	}

	return false
}

func (nap *newsArticleParser) parseArticleType(node *html.Node) {
	for _, attr := range node.Attr {
		if attr.Key == "content" {
			nap.art.Type = ArticleType(attr.Val)
			return
		}
	}
}
