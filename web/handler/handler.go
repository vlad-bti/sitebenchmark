package handler

import (
	"context"
	"net/http"
	"test/awesomeproject/service"
	"time"

	"github.com/gin-gonic/gin"
	"test/awesomeproject/web"
	"test/awesomeproject/web/bind"
)

func GetSitesData(ctx *gin.Context) {
	q := bind.Query{}
	if err := ctx.ShouldBind(&q); err != nil {
		ctx.JSON(http.StatusOK, web.NewResp(1001, err.Error()))
		return
	}

	gctx, cancel := context.WithTimeout(ctx, time.Duration(service.Mgr().BenchmarkCenterService().GetMaxResponseTime())*time.Second)
	defer cancel()

	resp := service.Mgr().BenchmarkCenterService().GetYandexSearchResult(gctx, q.Search)
	if resp.Error != nil {
		ctx.JSON(http.StatusOK, web.NewResp(1001, resp.Error.Error()))
		return
	}
	if len(resp.Items) == 0 {
		ctx.JSON(http.StatusOK, web.NewResp(1001, "empty url list"))
		return
	}

	maxUrl := service.Mgr().BenchmarkCenterService().GetMaxUrl()
	ch := make(chan int, 10)
	numGoroutine := 0
	for k, item := range resp.Items {
		if k >= maxUrl {
			break
		}
		if !service.Mgr().BenchmarkCenterService().HostResultExists(gctx, item.Host) {
			service.Mgr().BenchmarkCenterService().SetDefaultHostResult(gctx, item.Host)
			numGoroutine += 1
			go service.Mgr().BenchmarkCenterService().BenchmarkSite(gctx, ch, item.Host, item.Url)
		}
	}

	if numGoroutine != 0 {
		countResp := 0
		for {
			select {
			case <-ch:
			case <-gctx.Done():
			}
			countResp += 1
			if countResp >= numGoroutine || gctx.Err() != nil {
				break
			}
		}
	}
	data := service.Mgr().BenchmarkCenterService().GetOptimumConcurrency(ctx, resp.Items)

	ctx.JSON(http.StatusOK, web.NewResp(0, "", data))
}
