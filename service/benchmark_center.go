package service

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"test/awesomeproject/conf"
	"test/awesomeproject/util"
)

type BenchmarkCenter interface {
	GetMaxResponseTime() int
	GetMaxUrl() int
	GetYandexSearchResult(ctx context.Context, keyword string) util.ResponseStruct
	HostResultExists(ctx context.Context, host string) bool
	SetDefaultHostResult(ctx context.Context, host string)
	GetOptimumConcurrency(ctx context.Context, sites []util.ResponseItem) map[string]int
	BenchmarkSite(ctx context.Context, ch chan int, host string, url string)
	TestUrl(ctx context.Context, url string, numGoroutine int) int
}

var _ BenchmarkCenter = (*benchmarkCenter)(nil)

type benchmarkCenter struct {
	svcName             string
	maxResponseTime     int
	maxSiteResponseTime int
	maxUrl              int
	maxError            int
	numGoroutineList    []int
	cacheHosts          map[string]int
}

func NewBenchmarkCenter(serviceName string, cfg conf.Config) *benchmarkCenter {
	return &benchmarkCenter{
		svcName:             serviceName,
		maxResponseTime:     cfg.Handler.MaxResponseTime,
		maxSiteResponseTime: cfg.Handler.MaxSiteResponseTime,
		maxUrl:              cfg.Handler.MaxUrl,
		maxError:            cfg.Handler.MaxError,
		numGoroutineList:    cfg.Handler.NumGoroutineList,
		cacheHosts:          make(map[string]int),
	}
}

func (c *benchmarkCenter) GetMaxResponseTime() int {
	return c.maxResponseTime
}

func (c *benchmarkCenter) GetMaxUrl() int {
	return c.maxUrl
}

func (c *benchmarkCenter) GetYandexSearchResult(ctx context.Context, keyword string) util.ResponseStruct {
	var netClient = &http.Client{
		Timeout: time.Duration(c.maxSiteResponseTime) * time.Second,
	}

	resp, err := netClient.Get(fmt.Sprintf(util.BaseYandexURL, url.QueryEscape(keyword)))
	if err != nil {
		return util.ResponseStruct{Items: make([]util.ResponseItem, 0)}
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	return util.ParseYandexResponse(bodyBytes)
}

func (c *benchmarkCenter) GetOptimumConcurrency(ctx context.Context, sites []util.ResponseItem) map[string]int {
	ret := make(map[string]int, 0)
	for k, v := range sites {
		if k >= c.maxUrl {
			break
		}
		ret[v.Host] = -1
		if num, ok := c.cacheHosts[v.Host]; ok {
			ret[v.Host] = num
		}
	}
	return ret
}

func (c *benchmarkCenter) HostResultExists(ctx context.Context, host string) bool {
	if _, ok := c.cacheHosts[host]; ok {
		return true
	}
	return false
}

func (c *benchmarkCenter) SetDefaultHostResult(ctx context.Context, host string) {
	c.cacheHosts[host] = -1
}

func generate(abort <-chan struct{}, list []int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		i := 0
		var value int
		for {
			if len(list) > i {
				value = list[i]
			} else {
				value = -1
			}
			select {
			case ch <- value:
			case <-abort:
				return
			}
			i += 1
		}
	}()
	return ch
}

func (c *benchmarkCenter) BenchmarkSite(ctx context.Context, ch chan int, host string, url string) {
	abort := make(chan struct{})
	gen := generate(abort, c.numGoroutineList)

	var minNum int
	var maxNum int
	var numGoroutine int
	for {
		if minNum == numGoroutine {
			v := <-gen
			if v == -1 {
				if maxNum == 0 {
					numGoroutine *= 2
				} else {
					numGoroutine = minNum + (maxNum-minNum)/2
				}
			} else {
				numGoroutine = v
			}
		} else {
			numGoroutine = minNum + (maxNum-minNum)/2
		}
		numOk := c.TestUrl(ctx, url, numGoroutine)

		if numGoroutine-numOk <= c.maxError {
			c.cacheHosts[host] = numGoroutine
		}
		if ctx.Err() != nil || maxNum != 0 && maxNum-minNum < 5 {
			close(abort)
			break
		}
		if numGoroutine-numOk <= c.maxError {
			minNum = numGoroutine
		} else {
			maxNum = numGoroutine
		}
	}
	ch <- 1
}

func (c *benchmarkCenter) TestUrl(ctx context.Context, url string, numGoroutine int) int {
	gctx, cancel := context.WithTimeout(ctx, time.Duration(c.maxSiteResponseTime)*time.Second)
	defer cancel()

	ch := make(chan int, 10)
	defer close(ch)

	for i := 0; i < numGoroutine; i++ {
		go func(ch chan int, ctx context.Context) {
			var netClient = &http.Client{
				Timeout: time.Duration(c.maxSiteResponseTime) * time.Second,
			}

			resp, err := netClient.Get(url)
			if err != nil {
				if ctx.Err() == nil {
					ch <- -1
				}
				return
			}
			if ctx.Err() == nil {
				ch <- resp.StatusCode
			}
		}(ch, gctx)
	}
	numOk := 0
	countResp := 0

	for {
		select {
		case code := <-ch:
			if code == 200 {
				numOk += 1
			}
		case <-gctx.Done():
		}
		countResp += 1
		if countResp >= numGoroutine || gctx.Err() != nil {
			break
		}
	}

	return numOk
}
