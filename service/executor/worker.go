package executor

import (
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/xiaoxuxiansheng/xtimer/common/consts"
	"github.com/xiaoxuxiansheng/xtimer/common/model/vo"
	"github.com/xiaoxuxiansheng/xtimer/common/utils"
	taskdao "github.com/xiaoxuxiansheng/xtimer/dao/task"
	"github.com/xiaoxuxiansheng/xtimer/pkg/bloom"
	"github.com/xiaoxuxiansheng/xtimer/pkg/log"
	"github.com/xiaoxuxiansheng/xtimer/pkg/promethus"
	utils2 "github.com/xiaoxuxiansheng/xtimer/pkg/utils"
	"github.com/xiaoxuxiansheng/xtimer/pkg/xhttp"
)

type Worker struct {
	timerService *TimerService
	taskDAO      *taskdao.TaskDAO
	httpClient   *xhttp.JSONClient
	bloomFilter  *bloom.Filter
	reporter     *promethus.Reporter
}

func NewWorker(timerService *TimerService, taskDAO *taskdao.TaskDAO, httpClient *xhttp.JSONClient, bloomFilter *bloom.Filter, reporter *promethus.Reporter) *Worker {
	return &Worker{
		timerService: timerService,
		taskDAO:      taskDAO,
		httpClient:   httpClient,
		bloomFilter:  bloomFilter,
		reporter:     reporter,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.timerService.Start(ctx)
}

func (w *Worker) Work(ctx context.Context, timerIDUnixKey string) error {
	// log.InfoContextf(ctx, "executor_1 start: %v", time.Now())
	// defer func() {
	// 	log.InfoContextf(ctx, "executor_1 end: %v", time.Now())
	// }()
	// 拿到消息，查询一次完整的 timer 定义
	timerID, unix, err := utils.SplitTimerIDUnix(timerIDUnixKey)
	if err != nil {
		return err
	}

	if exist, err := w.bloomFilter.Exist(ctx, utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))), timerIDUnixKey); err != nil || exist {
		log.WarnContextf(ctx, "bloom filter check failed, start to check db, bloom key: %s, timerIDUnixKey: %s, err: %v, exist: %t", utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))), timerIDUnixKey, err, exist)
		// 查库判断定时器状态
		task, err := w.taskDAO.GetTask(ctx, taskdao.WithTimerID(timerID), taskdao.WithRunTimer(time.UnixMilli(unix)))
		if err == nil && task.Status != consts.NotRunned.ToInt() {
			// 重复执行的任务
			log.WarnContextf(ctx, "task is already executed, timerID: %d, exec_time: %v", timerID, task.RunTimer)
			return nil
		}
	}

	return w.executeAndPostProcess(ctx, timerID, unix)
}

func (w *Worker) executeAndPostProcess(ctx context.Context, timerID uint, unix int64) error {
	// 未执行，则查询 timer 完整的定义，执行回调
	timer, err := w.timerService.GetTimer(ctx, timerID)
	if err != nil {
		return fmt.Errorf("get timer failed, id: %d, err: %w", timerID, err)
	}

	// 定时器已经处于去激活态，则无需处理任务
	if timer.Status != consts.Enabled {
		log.WarnContextf(ctx, "timer has alread been unabled, timerID: %d", timerID)
		return nil
	}

	execTime := time.Now()
	resp, err := w.execute(ctx, timer)
	if err != nil {
		log.Errorf("exec cronjob failed,err:", err)
	}
	overTime := time.Now()
	costTime := overTime.Sub(execTime)
	// log.InfoContextf(ctx, "execute timer: %d, resp: %v, err: %v", timerID, resp, err)
	return w.postProcess(ctx, resp, err, timer.App, timerID, unix, execTime, int64(costTime))
}

func (w *Worker) execute(ctx context.Context, timer *vo.Timer) (map[string]interface{}, error) {
	var (
		resp map[string]interface{}
		err  error
	)
	log.Infof("params:%v", utils2.GetJsonStr(timer))
	switch strings.ToUpper(timer.NotifyHTTPParam.Method) {
	case nethttp.MethodGet:
		err = w.httpClient.Get(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, nil, &resp)
	case nethttp.MethodPatch:
		err = w.httpClient.Patch(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, timer.NotifyHTTPParam.Body, &resp)
	case nethttp.MethodDelete:
		err = w.httpClient.Delete(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, timer.NotifyHTTPParam.Body, &resp)
	case nethttp.MethodPost:
		err = w.httpClient.Post(ctx, timer.NotifyHTTPParam.URL, timer.NotifyHTTPParam.Header, timer.NotifyHTTPParam.Body, &resp)
	default:
		err = fmt.Errorf("invalid http method: %s, timer: %s", timer.NotifyHTTPParam.Method, timer.Name)
	}

	return resp, err
}

func (w *Worker) postProcess(ctx context.Context, resp map[string]interface{}, execErr error, app string, timerID uint, unix int64, execTime time.Time, costTime int64) error {
	go w.reportMonitorData(app, unix, execTime)
	if err := w.bloomFilter.Set(ctx, utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))), utils.UnionTimerIDUnix(timerID, unix), consts.BloomFilterKeyExpireSeconds); err != nil {
		log.ErrorContextf(ctx, "set bloom filter failed, key: %s, err: %v", utils.GetTaskBloomFilterKey(utils.GetDayStr(time.UnixMilli(unix))), err)
	}

	task, err := w.taskDAO.GetTask(ctx, taskdao.WithTimerID(timerID), taskdao.WithRunTimer(time.UnixMilli(unix)))
	if err != nil {
		return fmt.Errorf("get task failed, timerID: %d, runTimer: %d, err: %w", timerID, time.UnixMilli(unix), err)
	}

	respBody, _ := json.Marshal(resp)
	task.Output = string(respBody)
	task.CostTime = int(costTime)

	if execErr != nil {
		task.Status = consts.Failed.ToInt()
	} else {
		task.Status = consts.Successed.ToInt()
	}

	return w.taskDAO.UpdateTask(ctx, task)
}

func (w *Worker) reportMonitorData(app string, expectExecTimeUnix int64, acutalExecTime time.Time) {
	w.reporter.ReportExecRecord(app)
	// 上报毫秒
	w.reporter.ReportTimerDelayRecord(app, float64(acutalExecTime.UnixMilli()-expectExecTimeUnix))
}
