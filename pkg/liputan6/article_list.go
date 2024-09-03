package liputan6

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/tamboto2000/ivosight-crawler/pkg/htmlutil"
	"golang.org/x/net/html"
)

type Thumbnail struct {
	URl    string
	Width  int
	Height int
	Alt    string
}

type ArticleListItem struct {
	Type        ArticleType
	Title       string
	Link        string
	Thumbnail   Thumbnail
	Summary     string
	PublishedAt time.Time
	lpt6        *Liputan6
}

type articleListParser struct {
	list []*ArticleListItem
	item ArticleListItem
	lpt6 *Liputan6
}

func (parser *articleListParser) parseArticleList(ctx context.Context, lpt6 *Liputan6, link string) ([]*ArticleListItem, error) {
	parser.lpt6 = lpt6

	node, err := lpt6.commonReq(ctx, link)
	if err != nil {
		return nil, err
	}

	node = htmlutil.FindNode(node, htmlutil.DefaultFilter{
		Type: html.ElementNode,
		Data: "article",
		Attributes: []html.Attribute{
			{
				Key: "class",
				Val: "main",
			},
		},
	})

	if node == nil {
		return nil, nil
	}

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if parser.parseArticleItem(node) {
			return false, true
		}

		return true, true
	})

	return parser.list, nil
}

func (parser *articleListParser) parseArticleItem(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "article" {
		return false
	}

	isArticleListItem := false
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "articles--rows--item" {
			isArticleListItem = true
		}

		if attr.Key == "data-type" {
			parser.item.Type = ArticleType(attr.Val)
		}
	}

	if !isArticleListItem {
		return false
	}

	htmlutil.WalkSkipNodes(node, func(node *html.Node) (bool, bool) {
		if parser.parseArticleItemThumbnail(node) {
			return false, true
		}

		if node.Type == html.ElementNode && node.Data == "time" {
			parser.parseArticleItemPublishedDatetime(node)
			return false, true
		}

		if parser.parseArticleItemTitle(node) {
			return false, true
		}

		if parser.parseArticleItemSummary(node) {
			return false, true
		}

		return true, true
	})

	parser.item.lpt6 = parser.lpt6
	item := parser.item
	parser.list = append(parser.list, &item)
	parser.item = ArticleListItem{}

	return true
}

func (parser *articleListParser) parseArticleItemThumbnail(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "figure" {
		return false
	}

	isThumbnail := false
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "articles--rows--item__figure-thumbnail" {
			isThumbnail = true
			break
		}
	}

	if !isThumbnail {
		return false
	}

	htmlutil.WalkNodes(node, func(node *html.Node) bool {
		if node.Type == html.ElementNode && node.Data == "img" {
			var thumbnail Thumbnail
			for _, attr := range node.Attr {
				switch attr.Key {
				case "src":
					thumbnail.URl = attr.Val

				case "width":
					width, _ := strconv.Atoi(attr.Val)
					thumbnail.Width = width

				case "height":
					height, _ := strconv.Atoi(attr.Val)
					thumbnail.Height = height

				case "alt":
					thumbnail.Alt = attr.Val
				}
			}

			parser.item.Thumbnail = thumbnail

			return false
		}

		return true
	})

	return true
}

func (parse *articleListParser) parseArticleItemPublishedDatetime(node *html.Node) {
	for _, attr := range node.Attr {
		if attr.Key == "datetime" {
			dt, _ := time.Parse(time.RFC3339, attr.Val)
			parse.item.PublishedAt = dt

			break
		}
	}
}

func (parser *articleListParser) parseArticleItemTitle(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "h4" {
		return false
	}

	isTitle := false
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "articles--rows--item__title" {
			isTitle = true
			break
		}
	}

	if !isTitle {
		return false
	}

	htmlutil.WalkNodes(node, func(node *html.Node) bool {
		if node.Data == "a" {
			for _, attr := range node.Attr {
				switch attr.Key {
				case "href":
					parser.item.Link = attr.Val

				case "title":
					title := attr.Val
					title = strings.Replace(title, "\u00a0", " ", -1)

					parser.item.Title = title
				}
			}

			return false
		}

		return true
	})

	return true
}

func (parser *articleListParser) parseArticleItemSummary(node *html.Node) bool {
	if node.Type != html.ElementNode && node.Data != "div" {
		return false
	}

	isSummary := false
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "articles--rows--item__summary" {
			isSummary = true
			break
		}
	}

	if !isSummary {
		return false
	}

	var summary string
	htmlutil.WalkNodes(node, func(node *html.Node) bool {
		if node.Type == html.TextNode {
			summary += node.Data
		}

		return true
	})

	summary = strings.TrimSpace(summary)
	parser.item.Summary = summary

	return true
}
