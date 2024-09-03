// Package detik provides functionality for crawling Detik news portal (https://detik.com)
package detik

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

type Detik struct {
	cl *http.Client
}

func NewDetik(cl *http.Client) *Detik {
	if cl == nil {
		cl = http.DefaultClient
	}

	return &Detik{cl: cl}
}

func (dtk *Detik) ArticleListFromChannel(ctx context.Context, ch Channel) ([]*ArticleListItem, error) {
	node, err := dtk.commonReq(ctx, ch.baseURL)
	if err != nil {
		return nil, err
	}

	nif := new(newsItemFinder)
	list := nif.parseNewsItems(dtk, node)

	return list, nil
}

func (dtk *Detik) ArticleFromLink(ctx context.Context, link string) (Article, error) {
	nap := new(newsArticleParser)
	return nap.parseArticle(ctx, dtk, link)
}

func (dtk *Detik) commonReq(ctx context.Context, url string) (*html.Node, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
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

	res, err := dtk.cl.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	gzipread, err := gzip.NewReader(res.Body)
	if err != nil {
		return nil, err
	}

	// -- TESTING --

	// body, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	return nil, err
	// }

	// gzipread, err := gzip.NewReader(bytes.NewReader(body))
	// if err != nil {
	// 	return nil, err
	// }

	// f, err := os.Create("request_dump.html")
	// if err != nil {
	// 	return nil, err
	// }

	// defer f.Close()

	// gzipread1, err := gzip.NewReader(bytes.NewReader(body))
	// if err != nil {
	// 	return nil, err
	// }

	// body1, err := io.ReadAll(gzipread1)
	// if err != nil {
	// 	return nil, err
	// }

	// if _, err = f.Write(body1); err != nil {
	// 	return nil, err
	// }

	// -- TESTING --

	if res.StatusCode != 200 {
		if res.StatusCode == http.StatusNotFound {
			return nil, errors.New("news not found")
		}

		body, err := io.ReadAll(gzipread)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error fetching news items: %s", string(body))
	}

	node, err := html.Parse(gzipread)
	if err != nil {
		return nil, err
	}

	return node, nil
}
