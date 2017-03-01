package pttrss

import (
	"time"

	"github.com/gorilla/feeds"
	readability "github.com/julianshen/go-readability"
	"github.com/julianshen/gopttcrawler"
)

func GetRss(board string) (string, error) {
	ch, done := gopttcrawler.GetArticlesGo(board, 0)
	n := 20
	i := 0
	now := time.Now()

	feed := &feeds.Feed{
		Title:       board,
		Link:        &feeds.Link{Href: "https://www.ptt.cc/bbs/" + board},
		Description: board,
		Created:     now,
	}

	feed.Items = []*feeds.Item{}
	loc, _ := time.LoadLocation("Asia/Taipei")

	for article := range ch {
		if i >= n {
			done <- true
			break
		}
		i++

		article.Load()
		postTime, err := time.ParseInLocation("Mon Jan 2 15:04:05 2006", article.DateTime, loc)
		if err != nil {
			return "", err
		}

		doc, _ := readability.NewDocument(article.Content)
		item := &feeds.Item{
			Title:       article.Title,
			Link:        &feeds.Link{Href: article.Url},
			Description: doc.Text(),
			Author:      &feeds.Author{Name: article.Author},
			Created:     postTime,
		}

		feed.Items = append(feed.Items, item)
	}

	rss, err := feed.ToRss()
	if err != nil {
		return "", err
	}

	return rss, nil
}
