package detik

import (
	"context"
	"strconv"
	"time"

	"github.com/tamboto2000/ivosight-crawler/pkg/htmlutil"
	"golang.org/x/net/html"
)

type ArticleListItem struct {
	ArticleLink string
	ImageURL    string
	Title       string
	Subtitle    string
	PublishedAt time.Time
	dtk         *Detik
}

func (item *ArticleListItem) Article(ctx context.Context) (Article, error) {
	nap := new(newsArticleParser)
	return nap.parseArticle(ctx, item.dtk, item.ArticleLink)
}

const (
	newsItemFinderStateFindArticle = iota

	// states for finding news item image
	newsItemFinderStateFindImageDiv
	newsItemFinderStateFindImageUrl

	// states for finding news item title and subtitle
	newsItemFinderStateFindTitleH3
	newsItemFinderStateFindTitleLink
	newsItemFinderStateFindTitleText
	newsItemFinderStateFindSubtitleText

	// state for finding news published datetime
	newsItemFinderStateFindPublishedDatetimeDiv
	newsItemFinderStateFindPublishedDatetimeUnix

	// state for appending newly completed news item
	newsItemFinderStateAppendNew
)

type newsItemFinder struct {
	state      uint
	items      []*ArticleListItem
	onProgress ArticleListItem
	dtk        *Detik
}

func (nif *newsItemFinder) parseNewsItems(dtk *Detik, node *html.Node) []*ArticleListItem {
	node = htmlutil.FindNode(node, htmlutil.DefaultFilter{
		Type: html.ElementNode,
		Data: "div",
		Attributes: []html.Attribute{
			{
				Key: "id",
				Val: "indeks-container",
			},
		},
	})

	nif.dtk = dtk
	nif.state = newsItemFinderStateFindArticle

	if node == nil {
		return nil
	}

	htmlutil.WalkNodes(node, nif.walkNodes)

	return nif.items
}

func (nif *newsItemFinder) parseArticle(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "article" {
		return true
	}

	return false
}

func (nif *newsItemFinder) parseImageDiv(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "media__image" {
				return true
			}
		}
	}

	return false
}

func (nif *newsItemFinder) parseImageUrl(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "img" {
		for _, attr := range node.Attr {
			if attr.Key == "src" {
				nif.onProgress.ImageURL = attr.Val

				return true
			}
		}
	}

	return false
}

func (nif *newsItemFinder) parseTitleH3(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "h3" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "media__title" {
				return true
			}
		}
	}

	return false
}

func (nif *newsItemFinder) parseTitlelink(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "a" {
		var link string
		isArticleLink := false
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				link = attr.Val
				if isArticleLink {
					nif.onProgress.ArticleLink = link
					return true
				}
			}

			if attr.Key == "class" && attr.Val == "media__link" {
				isArticleLink = true
				if link != "" {
					nif.onProgress.ArticleLink = link
					return true
				}
			}
		}
	}

	return false
}

func (nif *newsItemFinder) parseSubtitleH2(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "h2" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "media__subtitle" {
				return true
			}
		}
	}

	return false
}

func (nif *newsItemFinder) parseTitleText(node *html.Node) bool {
	if node.Type == html.TextNode {
		nif.onProgress.Title = node.Data

		return true
	}

	return false
}

func (nif *newsItemFinder) parseSubtitleText(node *html.Node) bool {
	if node.Type == html.TextNode {
		nif.onProgress.Subtitle = node.Data

		return true
	}

	return false
}

func (nif *newsItemFinder) parseDatetimeDiv(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "media__date" {
				return true
			}
		}
	}

	return false
}

func (nif *newsItemFinder) parseDatetimeUnix(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "span" {
		for _, attr := range node.Attr {
			if attr.Key == "d-time" {
				unix, err := strconv.ParseInt(attr.Val, 10, 64)
				if err != nil {
					// We got the timestamp, but not a valid one,
					// the news article would be still valid
					return true
				}

				t := time.Unix(unix, 0)
				nif.onProgress.PublishedAt = t

				return true
			}
		}
	}

	return false
}

func (nif *newsItemFinder) walkNodes(node *html.Node) bool {
	switch nif.state {
	case newsItemFinderStateFindArticle:
		if nif.parseArticle(node) {
			nif.state = newsItemFinderStateFindImageDiv
		}

	case newsItemFinderStateFindImageDiv:
		if nif.parseImageDiv(node) {
			nif.state = newsItemFinderStateFindImageUrl
		}

	case newsItemFinderStateFindImageUrl:
		if nif.parseImageUrl(node) {
			nif.state = newsItemFinderStateFindTitleH3
		}

	case newsItemFinderStateFindTitleH3:
		if nif.parseSubtitleH2(node) {
			nif.state = newsItemFinderStateFindSubtitleText
			break
		}

		if nif.parseTitleH3(node) {
			nif.state = newsItemFinderStateFindTitleLink
		}

	case newsItemFinderStateFindTitleLink:
		if nif.parseTitlelink(node) {
			nif.state = newsItemFinderStateFindTitleText
		}

	case newsItemFinderStateFindTitleText:
		nif.parseTitleText(node)
		nif.state = newsItemFinderStateFindPublishedDatetimeDiv

	case newsItemFinderStateFindSubtitleText:
		nif.parseSubtitleText(node)
		nif.state = newsItemFinderStateFindTitleH3

	case newsItemFinderStateFindPublishedDatetimeDiv:
		if nif.parseDatetimeDiv(node) {
			nif.state = newsItemFinderStateFindPublishedDatetimeUnix
		}

	case newsItemFinderStateFindPublishedDatetimeUnix:
		if nif.parseDatetimeUnix(node) {
			nif.state = newsItemFinderStateAppendNew
		}

	case newsItemFinderStateAppendNew:
		item := nif.onProgress
		item.dtk = nif.dtk
		nif.items = append(nif.items, &item)
		nif.onProgress = ArticleListItem{}
		nif.state = newsItemFinderStateFindArticle
	}

	// Just in case if newsItemFinderStateAppendNew state is never reached and
	// we already encounter <article> tag, we will skip the item and
	// find other valid or parsable items
	if node.Type == html.ElementNode && node.Data == "article" {
		nif.state = newsItemFinderStateFindImageDiv
		nif.onProgress = ArticleListItem{}
	}

	return true
}
