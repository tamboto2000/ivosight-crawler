package crawler

import (
	"context"
	"log/slog"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/tamboto2000/ivosight-crawler/internal/config"
	"github.com/tamboto2000/ivosight-crawler/internal/models"
	"github.com/tamboto2000/ivosight-crawler/pkg/detik"
	"github.com/tamboto2000/ivosight-crawler/pkg/liputan6"
	"github.com/tamboto2000/ivosight-crawler/pkg/proxrotate"
	"github.com/tamboto2000/ivosight-crawler/pkg/random"
	"github.com/tamboto2000/ivosight-crawler/pkg/syncx"
)

type articleList struct {
	list []newsIndexItem
	rwmx sync.RWMutex
}

func (artlist *articleList) add(item newsIndexItem) {
	artlist.rwmx.Lock()
	defer artlist.rwmx.Unlock()

	sort.Slice(artlist.list, func(i, j int) bool {
		it := artlist.list[i].publishedAt
		jt := artlist.list[j].publishedAt

		return it.After(jt)
	})
}

func (artlist *articleList) get() (newsIndexItem, bool) {
	artlist.rwmx.RLock()
	defer artlist.rwmx.RUnlock()

	if len(artlist.list) == 0 {
		return newsIndexItem{}, false
	}

	item := artlist.list[0]
	artlist.list = artlist.list[1:]

	return item, true
}

var reqTimeout time.Duration = 30 * time.Second

type newsIndexItem struct {
	source      string
	link        string
	publishedAt time.Time
}

type Repository interface {
	// StoreArticle(ctx context.Context, article models.NewsArticle) error
	IsAlreadyExist(ctx context.Context, link string) (bool, error)
}

type NewsCrawler struct {
	routines    *syncx.Routines
	repo        Repository
	cfg         config.Crawler
	proxrot     *proxrotate.ProxyRotator
	articleList articleList
}

func (crawl *NewsCrawler) crawlNewsIndexes() error {
	for {
		interval := crawl.randomInterval()
		tc := time.After(interval)

		<-tc

		crawl.routines.WaitAvailable()
		crawl.routines.Go(crawl._crawlDetikIndex)
		crawl.routines.WaitAvailable()
		crawl.routines.Go(crawl._crawlLiputan6Index)
	}
}

func (crawl *NewsCrawler) _crawlDetikIndex() error {
	cl := http.DefaultClient
	cl.Timeout = reqTimeout
	crawl.proxrot.Rotate(cl)

	dtk := detik.NewDetik(cl)
	list, err := dtk.ArticleListFromChannel(context.Background(), detik.ChannelNews)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	for _, item := range list {
		exists, err := crawl.repo.IsAlreadyExist(context.Background(), item.ArticleLink)
		if err != nil {
			slog.Error(err.Error())
			return err
		}

		if !exists {
			crawl.articleList.add(newsIndexItem{
				source:      models.Detik,
				link:        item.ArticleLink,
				publishedAt: item.PublishedAt,
			})
		}
	}

	return nil
}

func (crawl *NewsCrawler) _crawlLiputan6Index() error {
	cl := http.DefaultClient
	cl.Timeout = reqTimeout
	crawl.proxrot.Rotate(cl)

	lpt := liputan6.NewLiputan6(cl)
	list, err := lpt.ArticleListFromIndex(context.Background())
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	for _, item := range list {
		exists, err := crawl.repo.IsAlreadyExist(context.Background(), item.Link)
		if err != nil {
			slog.Error(err.Error())
			return err
		}

		if !exists {
			crawl.articleList.add(newsIndexItem{
				source:      models.Liputan6,
				link:        item.Link,
				publishedAt: item.PublishedAt,
			})
		}
	}

	return nil
}

func (crawl *NewsCrawler) randomInterval() time.Duration {
	min := crawl.cfg.RandomRunIntervalRange[0]
	max := crawl.cfg.RandomRunIntervalRange[1]
	randnum := random.RandomNumRange(min, max)

	return time.Duration(randnum) * time.Second
}
