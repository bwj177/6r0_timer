package webserver

import (
	"fmt"
	"github.com/google/martian/log"
	"github.com/xiaoxuxiansheng/xtimer/middleware"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lxn/walk"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	gs "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"github.com/xiaoxuxiansheng/xtimer/common/conf"
)

type Server struct {
	sync.Once
	engine *gin.Engine

	timerApp *TimerApp
	taskApp  *TaskApp
	userApp  *UserApp

	userRouter  *gin.RouterGroup
	timerRouter *gin.RouterGroup
	taskRouter  *gin.RouterGroup
	mockRouter  *gin.RouterGroup

	confProvider *conf.WebServerAppConfProvider
}

func NewServer(timer *TimerApp, task *TaskApp, user *UserApp, confProvider *conf.WebServerAppConfProvider) *Server {
	s := Server{
		engine:       gin.Default(),
		timerApp:     timer,
		taskApp:      task,
		userApp:      user,
		confProvider: confProvider,
	}

	s.engine.Use(CrosHandler())
	s.timerRouter = s.engine.Group("api/timer/v1")
	s.taskRouter = s.engine.Group("api/task/v1")
	s.mockRouter = s.engine.Group("api/mock/v1")
	s.userRouter = s.engine.Group("api/user/v1")
	s.RegisterBaseRouter()
	s.RegisterMockRouter()
	s.RegisterUserRouter()
	s.RegisterTimerRouter()
	s.RegisterTaskRouter()
	s.RegisterMonitorRouter()
	return &s
}

func (s *Server) Start() {
	s.Do(s.start)
}

func (s *Server) start() {
	conf := s.confProvider.Get()
	go func() {
		if err := s.engine.Run(fmt.Sprintf(":%d", conf.Port)); err != nil {
			panic(err)
		}
	}()
}

func (s *Server) RegisterBaseRouter() {
	s.engine.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
}

func (s *Server) RegisterTimerRouter() {
	s.timerRouter.Use(middleware.JWTAuthMiddleware())
	s.timerRouter.GET("/def", s.timerApp.GetTimer)
	s.timerRouter.POST("/def", s.timerApp.CreateTimer)
	s.timerRouter.DELETE("/def", s.timerApp.DeleteTimer)
	s.timerRouter.POST("/update", s.timerApp.UpdateTimer)

	s.timerRouter.GET("/defs", s.timerApp.GetAppTimers)
	s.timerRouter.GET("/defsByName", s.timerApp.GetTimersByName)

	s.timerRouter.POST("/enable", s.timerApp.EnableTimer)
	s.timerRouter.POST("/unable", s.timerApp.UnableTimer)
}

func (s *Server) RegisterUserRouter() {
	s.userRouter.POST("/signup", s.userApp.SignUp)
	s.userRouter.POST("/login", s.userApp.Login)
}

func (s *Server) RegisterTaskRouter() {
	s.taskRouter.GET("/records", s.taskApp.GetTasks)
}

func (s *Server) RegisterMockRouter() {
	s.mockRouter.Any("/mock", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, struct {
			Word string `json:"word"`
		}{
			Word: "hello world!",
		})
	})

	s.mockRouter.POST("/test", func(ctx *gin.Context) {

		// 测试数据，用来标识定时任务已执行
		ShowTimePopup()
		ctx.JSON(http.StatusOK, map[string]string{"msg": "cron test exec success"})
	})

}

func (s *Server) RegisterMonitorRouter() {
	s.engine.Any("/metrics", func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})
}

func ShowTimePopup() {
	// 获取当前时间
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	// 构建消息内容
	message := fmt.Sprintf("现在的时间是：%s", currentTime)
	log.Infof("cur time cron test:%v", message)
	// 弹出消息框
	go func() { walk.MsgBox(nil, "当前时间", message, walk.MsgBoxIconInformation) }()

	// 模拟定时任务执行耗时
	var DelayTime = 200 * time.Millisecond
	var RangeTime = 500 * time.Millisecond
	randRange := int64(RangeTime)
	randDelay := DelayTime + time.Duration(rand.Int63()%randRange)
	time.Sleep(randDelay)
	return
}
