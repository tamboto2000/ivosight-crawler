// Package liputan6 provides functionality to crawl Liputan 6 news portal (https://liputan6.com)
package liputan6

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/net/html"
)

const UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36"

const liputanIndex = "https://www.liputan6.com/indeks"

type Liputan6 struct {
	cl *http.Client
}

func NewLiputan6(cl *http.Client) *Liputan6 {
	if cl == nil {
		cl = http.DefaultClient
	}

	return &Liputan6{cl: cl}
}

func (lpt6 *Liputan6) ArticleListFromLink(ctx context.Context, link string) ([]*ArticleListItem, error) {
	parser := new(articleListParser)
	return parser.parseArticleList(ctx, lpt6, link)
}

func (lpt6 *Liputan6) ArticleListFromIndex(ctx context.Context) ([]*ArticleListItem, error) {
	parser := new(articleListParser)
	return parser.parseArticleList(ctx, lpt6, liputanIndex)
}

func (lpt6 *Liputan6) ArticleFromLink(ctx context.Context, link string) (Article, error) {
	parser := new(articleParser)
	return parser.parseArticle(ctx, lpt6, &ArticleListItem{Link: link})
}

func (lpt6 *Liputan6) commonReq(ctx context.Context, url string) (*html.Node, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("accept-language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("cache-control", "max-age=0")
	req.Header.Set("priority", "u=0, i")
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", "\"Linux\"")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "none")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", UserAgent)

	res, err := lpt6.cl.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error requesting resource: got %d status code", res.StatusCode)
	}

	return html.Parse(res.Body)
}
