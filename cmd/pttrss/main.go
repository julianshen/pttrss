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

	var iTTL int
	if i, e := strconv.Atoi(ttl); e != nil {
		iTTL = 30
	} else {
		iTTL = i
	}

	log.Println(iTTL)

	router := gin.Default()
	cache := cache.New(time.Minute*time.Duration(iTTL), 30*time.Second)

	router.GET("/rss/:board", func(c *gin.Context) {
		board := c.Param("board")
		var rss string
		var e error
		var cacheItem = &CacheItem{}

		ifNonMatchStr := c.Request.Header.Get(IfNoneMatch)

		mutex.Lock()
		if val, existed := cache.Get(board); existed {
			log.Println("cached")
			var ok bool
			if cacheItem, ok = val.(*CacheItem); ok {
				if ifNonMatchStr != "" && ifNonMatchStr == cacheItem.etag {
					c.String(304, "Not modified")
					return
				}
			}
		} else {
			rss, e = pttrss.GetRss(board)

			if e == nil {
				etag := pttrss.Etag(rss)
				cacheItem.rss = rss
				cacheItem.etag = etag
				cache.Set(board, cacheItem, time.Minute*time.Duration(iTTL))
			}
		}
		mutex.Unlock()

		if e != nil {
			c.String(404, "Error to find %v:%v", board, e)
		} else {
			c.Header("Cache-Control", fmt.Sprintf("max-age=%d", iTTL*60))
			if cacheItem.etag != "" {
				c.Header("ETag", cacheItem.etag)
			}
			c.Data(200, "application/rss+xml", []byte(cacheItem.rss))
		}
	})

	router.Run(":" + port)
}
