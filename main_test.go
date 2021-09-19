package main

import (
	"context"
	"test/awesomeproject/conf"
	"test/awesomeproject/service"
	"test/awesomeproject/util"
	"testing"
	"time"
)

func TestYandexSearchResult(t *testing.T) {
	keyword := "playstation купить"

	var cfg conf.Config
	conf.ReadConfig(&cfg)

	err := service.Mgr().Register(
		service.NewBenchmarkCenter(svcBenchmarkCenter, cfg),
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	resp := service.Mgr().BenchmarkCenterService().GetYandexSearchResult(ctx, keyword)
	if resp.Error != nil {
		t.Fatal(resp.Error)
	}

	for k, v := range resp.Items {
		println(k, v.Host, v.Url)
	}
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

func TestYieldGenerator(t *testing.T) {
	abort := make(chan struct{})
	var cfg conf.Config
	conf.ReadConfig(&cfg)

	gen := generate(abort, cfg.Handler.NumGoroutineList)

	var minNum int
	var numGoroutine int
	for {
		if minNum == numGoroutine {
			v := <-gen
			if v == -1 {
				numGoroutine *= 2
			} else {
				numGoroutine = v
			}
		} else {
			numGoroutine = minNum + (numGoroutine-minNum)/2
		}

		numOk := numGoroutine
		if numGoroutine > 1000 {
			numOk = 0
		}

		println(minNum, numGoroutine)

		if numGoroutine-minNum < 10 {
			close(abort)
			break
		}
		if numGoroutine == numOk {
			minNum = numGoroutine
		}
	}
}

func TestUrl(t *testing.T) {
	url := "https://www.sentektechnologies.com/"
	numGoroutine := 70

	var cfg conf.Config
	conf.ReadConfig(&cfg)

	err := service.Mgr().Register(
		service.NewBenchmarkCenter(svcBenchmarkCenter, cfg),
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	resp := service.Mgr().BenchmarkCenterService().TestUrl(ctx, url, numGoroutine)

	println(numGoroutine, resp, url)
}

func TestBenchmarkSite(t *testing.T) {
	url := "https://www.sentektechnologies.com/"
	host := "www.sentektechnologies.com"

	//url = "http://127.0.0.1:8080/sites?search=bar"
	//host = "127.0.0.1:8080"

	var cfg conf.Config
	conf.ReadConfig(&cfg)

	err := service.Mgr().Register(
		service.NewBenchmarkCenter(svcBenchmarkCenter, cfg),
	)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan int, 10)
	ctx := context.Background()

	gctx, cancel := context.WithTimeout(ctx, time.Duration(service.Mgr().BenchmarkCenterService().GetMaxResponseTime())*time.Second)
	defer cancel()

	service.Mgr().BenchmarkCenterService().BenchmarkSite(gctx, ch, host, url)

	sites := make([]util.ResponseItem, 0)
	var item util.ResponseItem
	item.Url = url
	item.Host = host
	sites = append(sites, item)

	data := service.Mgr().BenchmarkCenterService().GetOptimumConcurrency(ctx, sites)
	for k, v := range data {
		println(k, v)
	}

	close(ch)
}
