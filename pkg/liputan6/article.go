package liputan6

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
	TextArticle  = "Article"
	VideoArticle = "Video"
	PhotoArticle = "Photo"
)

type ArticleContentType string

const (
	SectionTitle  ArticleContentType = "section-title"
	ParagraphText ArticleContentType = "paragraph-text"
	Image         ArticleContentType = "image"
	Video         ArticleContentType = "video"
	PublishedFrom ArticleContentType = "published-from"
)

type ArticleContent struct {
	Type ArticleContentType
	Data any
}

func (ct ArticleContent) String() string {
	str, ok := ct.Data.(string)
	if !ok {
		panic("not a string")
	}

	return str
}

func (ct ArticleContent) Image() ContentImage {
	img, ok := ct.Data.(ContentImage)
	if !ok {
		panic("not an image")
	}

	return img
}

type jsonScript struct {
	Author *struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"author,omitempty"`
}

type ArticleAuthor struct {
	Name       string
	ProfileURL string
}

type ContentImage struct {
	URL     string
	Title   string
	Caption string
	Alt     string
	Width   int
	Height  int
}

type RelatedArticle struct {
	Title       string
	ArticleLink string
	Thumbnail   ContentImage
}

type Article struct {
	Type            ArticleType
	Link            string
	Headline        string
	Description     string
	PublishedAt     time.Time
	UpdatedAt       time.Time
	Author          ArticleAuthor
	Contents        []ArticleContent
	RelatedArticles []RelatedArticle
}

type articleParser struct {
	article Article
}

func (parser *articleParser) parseArticle(ctx context.Context, lpt6 *Liputan6, item *ArticleListItem) (Article, error) {
	parser.article.Type = item.Type
	parser.article.Link = item.Link

	node, err := lpt6.commonReq(ctx, item.Link)
	if err != nil {
		return Article{}, err
	}

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if parser.isArticleTitle(node) {
			parser.parseArticleTitle(node)
			return false, true
		}

		if parser.parseArticleDescription(node) {
			return false, true
		}

		if parser.parseArticlePublishedDatetime(node) {
			return false, true
		}

		if parser.parseArticleUpdatedDatetime(node) {
			return false, true
		}

		if parser.article.Type == "" {
			parser.parseArticleType(node)
			return true, true
		}

		if parser.article.Type != "" {
			switch parser.article.Type {
			case TextArticle:

			}
		}

		switch parser.article.Type {
		case TextArticle:
			if parser.parseArticleMain(node) {
				return false, true
			}

		case PhotoArticle:
			if parser.parsePhotoArticleSlider(node) {
				return false, true
			}
		}

		if parser.parseJsonScript(node) {
			return false, true
		}

		return true, true
	})

	return parser.article, nil
}

func (parser *articleParser) isArticleTitle(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "title" && len(node.Attr) == 0 {
		return true
	}

	return false
}

func (parser *articleParser) parseArticleTitle(node *html.Node) {
	var title string
	htmlutil.WalkNodes(node, func(node *html.Node) bool {
		if node.Type == html.TextNode {
			title += node.Data
		}

		return true
	})

	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, "\u00a0", " ")
	parser.article.Headline = title
}

func (parser *articleParser) parseArticleDescription(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "meta" {
		return false
	}

	var desc string

	isDesc := false
	for _, attr := range node.Attr {
		if attr.Key == "name" && attr.Val == "description" {
			isDesc = true
		}

		if attr.Key == "content" {
			desc = attr.Val
		}
	}

	if !isDesc {
		return false
	}

	parser.article.Description = desc

	return true
}

func (parser *articleParser) parseArticlePublishedDatetime(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "meta" {
		return false
	}

	isDatetime := false
	for _, attr := range node.Attr {
		switch attr.Key {
		case "property":
			if attr.Val == "article:published_time" {
				isDatetime = true
			}

		case "content":
			dt, err := time.Parse(time.RFC3339, attr.Val)
			if err != nil {
				return false
			}

			parser.article.PublishedAt = dt
		}
	}

	return isDatetime
}

func (parser *articleParser) parseArticleUpdatedDatetime(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "meta" {
		return false
	}

	isDatetime := false
	for _, attr := range node.Attr {
		if attr.Key == "property" && attr.Val == "article:modified_time" {
			isDatetime = true
		}

		if attr.Key == "content" {
			dt, err := time.Parse(time.RFC3339, attr.Val)
			if err != nil {
				return false
			}

			parser.article.UpdatedAt = dt
		}
	}

	return isDatetime
}

func (parser *articleParser) parseArticleType(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "body" {
		return false
	}

	for _, attr := range node.Attr {
		if attr.Key == "class" {
			switch attr.Val {
			case "articles show category-news immersive":
				parser.article.Type = TextArticle
			case "articles show category-photo immersive":
				parser.article.Type = PhotoArticle
			}

			return true
		}
	}

	return true
}

func (parser *articleParser) parseArticleMain(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "article" {
		return false
	}

	isMain := false
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "hentry main" {
			isMain = true
			break
		}
	}

	if !isMain {
		return false
	}

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if parser.parseRelatedArticles(node) {
			return false, true
		}

		if parser.parseArticleMainContent(node) {
			return false, true
		}

		return true, true
	})

	return true
}

func (parser *articleParser) parseArticleMainContent(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "div" {
		return false
	}

	isMainContent := false
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "read-page--content" {
			isMainContent = true
			break
		}
	}

	if !isMainContent {
		return false
	}

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if parser.parseTopImage(node) {
			return false, true
		}

		if parser.parseMainContentBody(node) {
			return false, true
		}

		return true, true
	})

	return true
}

func (parser *articleParser) parseTopImage(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "div" {
		return false
	}

	isTopImage := false
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "read-page--top-media" {
			isTopImage = true
			break
		}
	}

	if !isTopImage {
		return false
	}

	node = htmlutil.FindNode(node, htmlutil.DefaultFilter{
		Type: html.ElementNode,
		Data: "figure",
		Attributes: []html.Attribute{
			{
				Key: "class",
				Val: "read-page--photo-gallery--item",
			},
		},
	})

	if node != nil {
		node = htmlutil.FindNode(node, htmlutil.DefaultFilter{
			Type: html.ElementNode,
			Data: "a",
			Attributes: []html.Attribute{
				{
					Key: "class",
					Val: "read-page--photo-gallery--item__link",
				},
			},
		})

		if node != nil {
			node := htmlutil.FindNode(node, htmlutil.DefaultFilter{
				Type: html.ElementNode,
				Data: "img",
			})

			if node != nil {
				var contImg ContentImage
				for _, attr := range node.Attr {
					switch attr.Key {
					case "data-src":
						contImg.URL = attr.Val

					case "width":
						width, _ := strconv.Atoi(attr.Val)
						contImg.Width = width

					case "height":
						height, _ := strconv.Atoi(attr.Val)
						contImg.Height = height

					case "alt":
						contImg.Caption = attr.Val
					}
				}

				parser.article.Contents = append(parser.article.Contents, ArticleContent{
					Type: Image,
					Data: contImg,
				})
			}
		}
	}

	return true
}

func (parser *articleParser) parseMainContentBody(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "div" {
		return false
	}

	isMainContentBody := false
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "article-content-body article-content-body_with-aside" {
			isMainContentBody = true
			break
		}
	}

	if !isMainContentBody {
		return false
	}

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if parser.parseMainContentBodyPageItem(node) {
			return false, true
		}

		return true, true
	})

	return true
}

var contentBodyPageRgx = regexp.MustCompile(`^article-content-body__item-page`)

func (parser *articleParser) parseMainContentBodyPageItem(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "div" {
		return false
	}

	isPageItem := false
	for _, attr := range node.Attr {
		if attr.Key == "class" {
			if contentBodyPageRgx.MatchString(attr.Val) {
				isPageItem = true
				break
			}
		}
	}

	if !isPageItem {
		return false
	}

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "div":
				isContImg := false
			LOOP:
				for _, attr := range node.Attr {
					if attr.Key == "class" {
						switch attr.Val {
						// skip ad component and promo
						case "advertisement-text", "promo promo-above", "promo promo-below":
							return false, true

						case "article-content-body__item-media":
							isContImg = true
							break LOOP
						}
					}
				}

				// Parse content image
				if isContImg {
					node = htmlutil.FindNode(node, htmlutil.DefaultFilter{
						Type: html.ElementNode,
						Data: "figure",
						Attributes: []html.Attribute{
							{
								Key: "class",
								Val: "read-page--photo-gallery--item",
							},
						},
					})

					if node != nil {
						var contImg ContentImage
						for _, attr := range node.Attr {
							switch attr.Key {
							case "data-image":
								contImg.URL = attr.Val
							case "data-title":
								contImg.Alt = attr.Val
							case "data-description":
								caption := attr.Val
								caption = strings.Trim(caption, "<>p/")
								caption = strings.ReplaceAll(caption, "\u00a0", " ")
								contImg.Caption = caption
							}
						}

						node = htmlutil.FindNode(node, htmlutil.DefaultFilter{
							Type: html.ElementNode,
							Data: "img",
						})

						if node != nil {
							for _, attr := range node.Attr {
								switch attr.Key {
								case "width":
									width, _ := strconv.Atoi(attr.Val)
									contImg.Width = width

								case "height":
									height, _ := strconv.Atoi(attr.Val)
									contImg.Height = height
								}
							}
						}

						parser.article.Contents = append(parser.article.Contents, ArticleContent{
							Type: Image,
							Data: contImg,
						})
					}

					return false, true
				}

			// Parse section title
			case "h2":
				isSectionTitle := false
				for _, attr := range node.Attr {
					if attr.Key == "class" && attr.Val == "article-content-body__item-title" {
						isSectionTitle = true
						break
					}
				}

				if !isSectionTitle {
					return false, true
				}

				var sectionTitle string
				htmlutil.WalkNodes(node, func(node *html.Node) bool {
					if node.Type == html.TextNode {
						sectionTitle += node.Data
					}

					return true
				})

				sectionTitle = strings.TrimSpace(sectionTitle)
				parser.article.Contents = append(parser.article.Contents, ArticleContent{
					Type: SectionTitle,
					Data: sectionTitle,
				})

				return false, true

			// Parse paragraph
			case "p":
				if len(node.Attr) != 0 {
					return false, true
				}

				var paragraph string
				htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
					switch node.Type {
					case html.ElementNode:
						if node.Data == "p" {
							return true, true
						}

						var buff bytes.Buffer
						html.Render(&buff, node)
						paragraph += buff.String()

						return false, true

					case html.TextNode:
						txt := node.Data
						paragraph += txt
					}

					return true, true
				})

				paragraph = strings.ReplaceAll(paragraph, "\u00a0", " ")
				if paragraph != "" && paragraph != " " {
					parser.article.Contents = append(parser.article.Contents, ArticleContent{
						Type: ParagraphText,
						Data: paragraph,
					})
				}

				return false, true
			}
		}

		return true, true
	})

	return false
}

func (parser *articleParser) parseRelatedArticles(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "div" {
		return false
	}

	isRelatedArticles := false
	for _, attr := range node.Attr {
		if attr.Key == "id" && attr.Val == "related-news" {
			isRelatedArticles = true
			break
		}
	}

	if !isRelatedArticles {
		return false
	}

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if node.Type == html.ElementNode && node.Data == "div" {
			isItem := false
			for _, attr := range node.Attr {
				if attr.Key == "class" && attr.Val == "relateds-slow--lattice--item" {
					isItem = true
					break
				}
			}

			if !isItem {
				return true, true
			}

			var related RelatedArticle
			for _, attr := range node.Attr {
				if attr.Key == "data-title" {
					related.Title = attr.Val
					break
				}
			}

			node = htmlutil.FindNode(node, htmlutil.DefaultFilter{
				Type: html.ElementNode,
				Data: "figure",
			})

			if node != nil {
				htmlutil.WalkNodes(node, func(node *html.Node) bool {
					switch node.Data {
					case "a":
						for _, attr := range node.Attr {
							if attr.Key == "href" {
								related.ArticleLink = attr.Val
								return true
							}
						}

					case "img":
						for _, attr := range node.Attr {
							switch attr.Key {
							case "width":
								width, _ := strconv.Atoi(attr.Val)
								related.Thumbnail.Width = width

							case "height":
								height, _ := strconv.Atoi(attr.Val)
								related.Thumbnail.Height = height

							case "data-src":
								related.Thumbnail.URL = attr.Val
							}
						}
					}

					return true
				})

				parser.article.RelatedArticles = append(parser.article.RelatedArticles, related)
			}

			return false, true
		}

		return true, true
	})

	return true
}

func (parser *articleParser) parsePhotoArticleSlider(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "div" {
		return false
	}

	isSlider := false
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "read-page--photo-tag--slider__top js-top-slider" {
			isSlider = true
			break
		}
	}

	if !isSlider {
		return false
	}

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if node.Type == html.ElementNode && node.Data == "figure" {
			var contImg ContentImage

			for _, attr := range node.Attr {
				switch attr.Key {
				case "data-image":
					contImg.URL = attr.Val

				case "data-description":
					contImg.Caption = attr.Val

				case "data-title":
					contImg.Title = attr.Val
				}
			}

			node := htmlutil.FindNode(node, htmlutil.DefaultFilter{
				Type: html.ElementNode,
				Data: "img",
			})

			if node != nil {
				for _, attr := range node.Attr {
					switch attr.Key {
					case "width":
						width, _ := strconv.Atoi(attr.Val)
						contImg.Width = width

					case "height":
						height, _ := strconv.Atoi(attr.Val)
						contImg.Height = height
					}
				}
			}

			parser.article.Contents = append(parser.article.Contents, ArticleContent{
				Type: Image,
				Data: contImg,
			})

			return false, true
		}

		return true, true
	})

	return true
}

func (parser *articleParser) parseJsonScript(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "script" {
		return false
	}

	checkId := false
	checkType := false
	for _, attr := range node.Attr {
		switch attr.Key {
		case "id":
			if attr.Val != "rich-card" {
				return false
			}

			checkId = true

		case "type":
			if attr.Val != "application/ld+json" {
				return false
			}

			checkType = true
		}
	}

	if !checkId || !checkType {
		return false
	}

	var scripts []jsonScript
	var txt string
	htmlutil.WalkNodes(node, func(node *html.Node) bool {
		if node.Type == html.TextNode {
			txt += node.Data
		}

		return true
	})

	if err := json.Unmarshal([]byte(txt), &scripts); err != nil {
		return false
	}

	for _, s := range scripts {
		if s.Author != nil {
			parser.article.Author = ArticleAuthor{
				Name:       s.Author.Name,
				ProfileURL: s.Author.URL,
			}
		}
	}

	return true
}
