# 常用结构
**用户结构**
```go
// 用户账号结构体
type TableUser struct {
	Id       int64			// 用户 ID
	Name     string			// 用户名
	Password string			// 密码
}

// 用户信息结构体
type User struct {
	Id             int64  `json:"id,omitempty"`					// 用户 ID
	Name           string `json:"name,omitempty"`				// 用户名
	FollowCount    int64  `json:"follow_count"`					// 查询对象的关注数
	FollowerCount  int64  `json:"follower_count"`				// 查询对象的粉丝数
	IsFollow       bool   `json:"is_follow"`					// 当前用户是否关注该查询对象
	TotalFavorited int64  `json:"total_favorited,omitempty"`
	FavoriteCount  int64  `json:"favorite_count,omitempty"`
}
```
**响应报文**
```go
// 基础响应报文 (状态码, 状态信息)
type Response struct {
	StatusCode int32  `json:"status_code"`			// 状态码
	StatusMsg  string `json:"status_msg,omitempty"`	// 状态信息
}

// 用户登录响应报文
type UserLoginResponse struct {
	Response
	UserId int64  `json:"user_id,omitempty"`	// 用户 ID
	Token  string `json:"token"`				// token
}

// 用户信息响应报文
type UserResponse struct {
	Response
	User service.User `json:"user"`
}
```

# 用户登录功能
```go
// 1.1 Gin 路由组监听用户登录事件
apiRouter.POST("/user/login/", controller.Login)
// 1.2 从 URL 中获取用户名和密码
username := c.Query("username")
password := c.Query("password")
// 1.3 将密码转换为 sha256 处理后的加密密码
encoderPassword := service.EnCoder(password)
// 1.4 通过用户名找到用户信息对象 User(用户ID, 用户名, 密码)
u := usi.GetTableUserByUsername(username)
// 1.5 将数据库中的加密密码与用户提供的加密密码比对, 看是否一致
if encoderPassword == u.Password { ... }
// 1.6 根据用户名生成一个 token 字符串, 其中包含着一些用户信息 (用户名, 用户ID, 过期时间...)
token := service.GenerateToken(username)
// 1.7 生成一个登录响应返回给客户端
c.JSON(http.StatusOK, UserLoginResponse{
	Response: Response{StatusCode: 0},
	UserId: u.Id,
	Token: token,
})
```

# 用户注册功能
```go
// 2.1 Gin 路由组监听用户注册事件
apiRouter.POST("/user/register/", controller.Register)
// 2.2 从 URL 中获取用户名和密码
username := c.Query("username")
password := c.Query("password")
// 2.3 根据用户名从数据库里搜索, 判断该用户是否已存在
u := usi.GetTableUserByUsername(username)
if username == u.Name { ... }
// 2.4 将新用户的信息添加到数据库中, 其中密码存储的是加密密码
usi.InsertTableUser(&newUser)
// 2.5 根据用户名生成一个 token 字符串, 其中包含着一些用户信息 (用户名, 用户ID, 过期时间...)
token := service.GenerateToken(username)
// 2.6 生成一个用户登录响应报文返回给客户端 (注册好自动登录)
c.JSON(http.StatusOK, UserLoginResponse{
	Response: Response{StatusCode: 0},
	UserId: u.Id,
	Token: token,
})
```

# 获取用户信息功能
```go
// 3.1 Gin 路由组监听获取用户信息事件
apiRouter.GET("/user/", jwt.Auth(), controller.UserInfo)
// 3.2 从 URL 中获取用户 ID
user_id := c.Query("user_id")
// 3.3 根据用户 ID 从数据库中获取用户信息对象
u, err := usi.GetUserById(id)
// 3.4 返回用户信息响应报文
c.JSON(http.StatusOK, UserResponse{
	Response: Response{StatusCode: 0},
	User: u,
})
```

# 相关知识
[JWT token](https://www.ruanyifeng.com/blog/2018/07/json_web_token-tutorial.html)