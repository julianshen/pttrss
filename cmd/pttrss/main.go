package main

import (
	"log"
	"os"
	"sync"
	"time"

	"strconv"

	"fmt"

	"github.com/julianshen/pttrss"
	cache "github.com/robfig/go-cache"
	"gopkg.in/gin-gonic/gin.v1"
)

const (
	IfNoneMatch = "If-None-Match"
)

type CacheItem struct {
	rss  string
	etag string
}

var mutex = &sync.RWMutex{}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	ttl := os.Getenv("TTL") //TTL in minute

	var iTTL time.Duration
	if i, e := strconv.Atoi(ttl); e != nil {
		iTTL = time.Duration(30)
	} else {
		iTTL = time.Duration(i)
	}

	log.Println(iTTL)

	router := gin.Default()
	cache := cache.New(iTTL*time.Minute, 30*time.Second)

	router.GET("/rss/:board", func(c *gin.Context) {
		board := c.Param("board")
		var rss string
		var e error
		var cacheItem CacheItem

		ifNonMatchStr := c.Request.Header.Get(IfNoneMatch)

		mutex.Lock()
		if val, existed := cache.Get(board); existed {
			log.Println("cached")
			if cacheItem, ok := val.(*CacheItem); ok {
				if ifNonMatchStr != "" && ifNonMatchStr == cacheItem.etag {
					c.String(304, "Not modified")
					return
				}
			}
		} else {
			rss, e = pttrss.GetRss(board)
			log.Println("no cache")

			if e == nil {
				etag := pttrss.Etag(rss)
				cacheItem.rss = rss
				cacheItem.etag = etag
				cache.Set(board, &cacheItem, iTTL*time.Minute)
			}
		}
		mutex.Unlock()

		if e != nil {
			c.String(404, "Error to find %v:%v", board, e)
		} else {
			c.Header("Cache-Control", fmt.Sprintf("max-age=%v", (iTTL*time.Minute).Seconds))
			if cacheItem.etag != "" {
				c.Header("ETag", cacheItem.etag)
			}
			c.Data(200, "application/rss+xml", []byte(cacheItem.rss))
		}
	})

	router.Run(":" + port)
}
