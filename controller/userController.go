package controller

import (
	"TikTok/dao"
	"TikTok/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

// 基础响应报文 (状态码, 状态信息)
type Response struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

// 用户登录响应报文
type UserLoginResponse struct {
	Response
	UserId int64  `json:"user_id,omitempty"`
	Token  string `json:"token"`
}

// 用户信息响应报文
type UserResponse struct {
	Response
	User service.User `json:"user"`
}

// 用户注册事件
func Register(c *gin.Context) {
	// 2.2 从 URL 中获取用户名和密码
	username := c.Query("username")
	password := c.Query("password")
	usi := service.UserServiceImpl{}
	// 2.3 根据用户名从数据库里搜索, 判断该用户是否已存在
	u := usi.GetTableUserByUsername(username)
	if username == u.Name {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "User already exist"},
		})
	} else {
		newUser := dao.TableUser{
			Name:     username,
			Password: service.EnCoder(password),
		}
		// 2.4 将新用户的信息添加到数据库中, 其中密码存储的是加密密码
		if usi.InsertTableUser(&newUser) != true {
			println("Insert Data Fail")
		}
		u := usi.GetTableUserByUsername(username)
		// 2.5 根据用户名生成一个 token 字符串, 其中包含着一些用户信息 (用户名, 用户ID, 过期时间...)
		token := service.GenerateToken(username)
		log.Println("注册返回的id: ", u.Id)
		// 2.6 生成一个登录响应返回给客户端 (注册好自动登录)
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 0},
			UserId:   u.Id,
			Token:    token,
		})
	}
}

// 用户登录函数
func Login(c *gin.Context) {
	// 1.2 从 URL 中获取用户名和密码
	username := c.Query("username")
	password := c.Query("password")
	// 1.3 将密码转换为 sha256 处理后的加密密码
	encoderPassword := service.EnCoder(password)
	println(encoderPassword)
	usi := service.UserServiceImpl{}
	// 1.4 通过用户名在数据库中查找, 并返回用户信息对象 User(用户ID, 用户名, 密码)
	u := usi.GetTableUserByUsername(username)
	// 1.5 将数据库中的加密密码与用户提供的加密密码比对, 看是否一致
	if encoderPassword == u.Password {
		// 1.6 根据用户名生成一个 token 字符串, 其中包含着一些用户信息 (用户名, 用户ID, 过期时间...)
		token := service.GenerateToken(username)
		// 1.7 生成一个登录响应返回给客户端
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 0},
			UserId:   u.Id,
			Token:    token,
		})
	} else {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "Username or Password Error"},
		})
	}
}

// 获取用户信息
func UserInfo(c *gin.Context) {
	// 3.2 从 URL 中获取用户ID
	user_id := c.Query("user_id")
	id, _ := strconv.ParseInt(user_id, 10, 64)
	usi := service.UserServiceImpl{
		FollowService: &service.FollowServiceImp{},
		LikeService:   &service.LikeServiceImpl{},
	}
	// 3.3 根据用户 ID 从数据库中获取用户信息对象
	if u, err := usi.GetUserById(id); err != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 1, StatusMsg: "User Doesn't Exist"},
		})
	} else {
		// 3.4 返回用户信息响应报文
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0},
			User:     u,
		})
	}
}
