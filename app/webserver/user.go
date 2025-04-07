package webserver

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xiaoxuxiansheng/xtimer/common/model/vo"
	service "github.com/xiaoxuxiansheng/xtimer/service/webserver"
	"net/http"
)

type UserApp struct {
	service userService
}

func NewUserApp(service *service.UserService) *UserApp {
	return &UserApp{service: service}
}

func (t *UserApp) SignUp(c *gin.Context) {
	var req vo.SignUpReq
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[SignUp] bind req failed, err: %v", err)))
		return
	}

	err := t.service.SignUp(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusOK, vo.NewSignUpResp(vo.NewCodeMsg(-1, err.Error())))
	}
	c.JSON(http.StatusOK, vo.NewSignUpResp(vo.NewCodeMsgWithMsg("注册成功")))
}

func (t *UserApp) Login(c *gin.Context) {
	var req vo.LoginReq
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, vo.NewCodeMsg(-1, fmt.Sprintf("[Login] bind req failed, err: %v", err)))
		return
	}

	token, err := t.service.Login(c.Request.Context(), &req)
	if err != nil || token == "" {
		c.JSON(http.StatusOK, vo.NewCodeMsg(-1, "密码错误，请重试"))
		return
	}
	c.JSON(http.StatusOK, vo.NewLoginResp(token, vo.NewCodeMsgWithMsg("登录成功")))
}

type userService interface {
	SignUp(ctx context.Context, req *vo.SignUpReq) error
	Login(ctx context.Context, req *vo.LoginReq) (string, error)
}
