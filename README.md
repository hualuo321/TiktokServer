# 常用结构
**用户结构**
```go
// 用户基础信息结构体
type TableUser struct {
	Id       int64			// 用户 ID
	Name     string			// 用户名
	Password string			// 密码
}

// 用户详细信息信息结构体
type User struct {
	Id             int64  `json:"id,omitempty"` 			// 用户 ID
	Name           string `json:"name,omitempty"`			// 用户名
	FollowCount    int64  `json:"follow_count"` 			// 查询对象的关注数
	FollowerCount  int64  `json:"follower_count"`			// 查询对象的粉丝数
	IsFollow       bool   `json:"is_follow"`     			// 登录用户是否关注该查询对象
	TotalFavorited int64  `json:"total_favorited,omitempty"`// 查询用户的总被点赞量
	FavoriteCount  int64  `json:"favorite_count,omitempty"`	// 查询用户点赞了多少其他视频
}

// 视频表结构体
type TableVideo struct {
	Id          int64 `json:"id"`		// 视频 ID
	AuthorId    int64			// 作者 ID
	PlayUrl     string `json:"play_url"`	// 视频存放地址
	CoverUrl    string `json:"cover_url"`	// 封面存放地址
	PublishTime time.Time			// 发布事件
	Title       string `json:"title"` 	// 视频名称
}

// 视频信息结构体
type Video struct {
	dao.TableVideo
	Author        User  `json:"author"`				// 作者
	FavoriteCount int64 `json:"favorite_count"`		// 点赞量
	CommentCount  int64 `json:"comment_count"`		// 评论量
	IsFavorite    bool  `json:"is_favorite"`		// 登录用户对该视频是否点赞
}
```

**响应报文**
```go
// 基础响应 (状态码, 状态信息)
type Response struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

// 用户登录响应
type UserLoginResponse struct {
	Response
	UserId int64  `json:"user_id,omitempty"`
	Token  string `json:"token"`
}

// 用户信息响应
type UserResponse struct {
	Response
	User service.User `json:"user"`
}

// 点赞响应
type likeResponse struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

// 获取点赞列表响应
type GetFavouriteListResponse struct {
	StatusCode int32           `json:"status_code"`
	StatusMsg  string          `json:"status_msg,omitempty"`
	VideoList  []service.Video `json:"video_list,omitempty"`
}
```

# 功能介绍

## 用户登录功能
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

## 用户注册功能
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

## 获取用户信息功能
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

## 发布视频功能
```go
// 4.1 Gin 路由组监听发布视频事件
apiRouter.POST("/publish/action/", jwt.AuthBody(), controller.Publish)
// 4.2 获取请求中的表单数据, 用户 ID, 视频信息
data, err := c.FormFile("data")
userId, _ := strconv.ParseInt(c.GetString("userId"), 10, 64)
title := c.PostForm("title")
// 4.3 将视频数据在 FTP 文件服务器中, 通过 ffmpeg 截取视频封面
err = dao.VideoFTP(file, videoName)
err := Ffmpeg(f.VideoName, f.ImageName)
// 4.4 将视频信息, 封面信息放在数据库中
err = dao.Save(videoName, imageName, userId, title)
// 4.5 返回视频发布的响应报文
c.JSON(http.StatusOK, Response{
	StatusCode: 0,
	StatusMsg:  "uploaded successfully",
})
```

## 获取已发布视频功能
```go
// 5.1 Gin 路由组监听获取已发布视频事件
apiRouter.GET("/publish/list/", jwt.Auth(), controller.PublishList)
// 5.2 获取查询用户 ID, 登录用户 ID
user_Id, _ := c.GetQuery("user_id")
curId, _ := strconv.ParseInt(c.GetString("userId"), 10, 64)
// 5.3 根据查询用户 ID 获取他发布视频的视频列表
list, err := videoService.List(userId, curId)
data, err := dao.GetVideosByAuthorId(userId)
result := Db.Where(&TableVideo{AuthorId: authorId}).Find(&data)
// 5.4 根据视频基础信息组装视频详细信息 (协程)
err = videoService.copyVideos(&result, &data, curId)
videoService.creatVideos(&video, &temp, userId)
go func() { video.Author, err = videoService.GetUserByIdWithCurId(data.AuthorId, userId) }
go func() { video.FavoriteCount, err = videoService.FavouriteCount(data.Id) }
go func() { video.CommentCount, err = videoService.CountFromVideoId(data.Id) }
go func() { video.IsFavorite, err = videoService.IsFavourite(video.Id, userId) }
// 5.5 将获取到的视频详细信息返回给客户端
c.JSON(http.StatusOK, VideoListResponse{
	Response:  Response{StatusCode: 0},
	VideoList: list,
})
```

## 拉取视频列表到首页功能
```go
// 6.1 Gin 路由组监听拉取视频列表事件
apiRouter.GET("/feed/", jwt.AuthWithoutLogin(), controller.Feed)
// 6.2 获取请求中的参数信息 (最近时间, curID)
lastTime := c.Query("latest_time")
curId, _ := strconv.ParseInt(c.GetString("userId"), 10, 64)
// 6.3 根据用户 ID 和最近时间拉取视频流
feed, nextTime, err := videoService.Feed(lastTime, curId)
tableVideos, err := dao.GetVideosByLastTime(lastTime)
result := Db.Where("publish_time<?", lastTime).Order("publish_time desc").Limit(config.VideoCount).Find(&videos)
// 6.4 将视频的基本信息转化为视频的详细信息
err = videoService.copyVideos(&videos, &tableVideos, curId)
videoService.creatVideo(&video, &temp, curId)
go func() { video.Author, err = videoService.GetUserByIdWithCurId(data.AuthorId, userId) }
go func() { video.FavoriteCount, err = videoService.FavouriteCount(data.Id) }
go func() { video.CommentCount, err = videoService.CountFromVideoId(data.Id) }
go func() { video.IsFavorite, err = videoService.IsFavourite(video.Id, userId) }
// 6.5 返回拉取视频结果
c.JSON(http.StatusOK, FeedResponse{
    Response:  Response{StatusCode: 0},
    VideoList: feed,
    NextTime:  nextTime.Unix(),
})
```

## 

## 根据登录用户 ID 和查询用户 ID, 获取查询用户的详细信息
```go
// 1. 根据登录用户 ID 和查询用户 ID, 获取查询用户的详细信息
user = UserServer.GetUserByIdWithCurId(curID int64, userID int64)
// 2. 初始化空的用户详细信息结构体
user = User { ... }
// 3. 获取用户基本信息, 赋值给初始化空的结构体
tableUser = UserServer.GetTableUserById(userID int64)
tableUser = UserDao.GetTableUserById(userID int64)
Db.Where("id = ?", id).First(&tableUser)
# 赋值操作
user.name = tableUser.name ...
// 4. 获取查询用户的关注数, 赋值给初始化空的结构体
followCnt = FollowServer.GetFollowingCnt(userID int64)
# 检查 Redis 中是否存在该记录
followCnt = redis.RdbFollowing.SCard(redis.Ctx, strconv.Itoa(int(userId))).Result()
# 如果 Redis 中存在， 则更新过期时间, 返回
redis.RdbFollowing.Expire(redis.Ctx, strconv.Itoa(int(userId)), config.ExpireTime)
# 如果 Redis 中不存在， 则从数据库中查找符合条件的列表
ids = FollowDao.GetFollowingIds(userID int64)
Db.Model(Follow{}).Where("follower_id = ?", userId).Pluck("user_id", &ids)
# 将查询到的内容放到缓存中
go addFollowingToRedis(userId int, ids []int64)
followCnt = len(ids)
user.FollowCnt = followCnt
// 5. 获取查询用户的粉丝数, 赋值给初始化空的结构体
followerCnt = FollowServer.GetFollowerCnt(userID int64)
# 检查 Redis 中是否存在该记录
followerCnt = redis.RdbFollowers.SCard(redis.Ctx, strconv.Itoa(int(userId))).Result()
# 如果 Redis 中存在， 则更新过期时间, 返回
redis.RdbFollowers.Expire(redis.Ctx, strconv.Itoa(int(userId)), config.ExpireTime)
# 如果 Redis 中不存在， 则从数据库中查找符合条件的列表
ids = FollowDao.GetFollowersIds(userID int64)
Db.Model(Follow{}).Where("user_id = ?", userId).Where("cancel = ?", 0).Pluck("follower_id", &ids)
# 将查询到的内容放到缓存中
go addFollowersToRedis(int(userId), ids)
user.FollowerCnt = followerCnt
// 6. 判断登录使用是否关注查询用户
isfollow = FollowServer.IsFollowing(curID int64, userID int64)
# 检查 Redis 中是否存在该记录
flag = redis.RdbFollowingPart.SIsMember(redis.Ctx, strconv.Itoa(int(userId)), targetId).Result()
# 如果 Redis 中存在， 则更新过期时间, 返回
redis.RdbFollowingPart.Expire(redis.Ctx, strconv.Itoa(int(userId)), config.ExpireTime)
# 如果 Redis 中不存在， 则从数据库中查找关系
relation = FellowDao.FindRelation(curID, userID)
# 将查询到的内容放到缓存中
go addRelationToRedis(int(curID), int(userID))
user.isfollow = isfollow
// 7. 返回获取到的用户信息
return user
```

## 根据视频 ID 获取视频的点赞数量
```go
favoriteCnt = VideoServer.FavouriteCount(tableVideo.ID)
# 判断 Redis 中是否存在记录
n = redis.RdbLikeVideoId.Exists(redis.Ctx, strVideoId).Result()
# 如果 Redis 中存在， 则返回
count = redis.RdbLikeVideoId.SCard(redis.Ctx, strVideoId).Result()
# 如果 Redis 中不存在， 添加数据到缓存中
redis.RdbLikeVideoId.Del(redis.Ctx, strVideoId)
# 设置过期时间
redis.RdbLikeVideoId.Expire(redis.Ctx, strVideoId, time.Duration(config.OneMonth)*time.Second).Result()
# 根据 videoID 获取点赞的 userID
userID[] = GetLikeUserIdList(videoId int64)
Db.Model(Like{}).Where(map[string]interface{}{"video_id": videoId, "cancel": config.IsLike}).Pluck("user_id", &likeUserIdList)
# 将相关信息写入缓存, 重新从缓存中读取数据
redis.RdbLikeVideoId.SAdd(redis.Ctx, strVideoId, likeUserId)
count = redis.RdbLikeVideoId.SCard(redis.Ctx, strVideoId).Result()
```









# 中间件
## JWT 鉴权模块
**Auth()**
```go
// 首先会获取token，检查用户携带的token是否正确，
auth := context.Query("token")
token, err := parseToken(auth)
// 如果token正确，则将用登录者的用户ID放入上下文中，并放行
context.Set("userId", token.Id)
context.Next()
// 如果token不正确，则终止
context.Abort()
```
**AuthWithoutLogin()**
```go
// 在未登录情况下，如果携带token，则会解析token检查是否正确
auth := context.Query("token")
token, err := parseToken(auth)
// 如果没有携带token，则userID默认为0，并放行
userId = "0"
// 如果token正确，则将用登录者的userID放入上下文中，并放行
context.Set("userId", userId)
context.Next()
```
****

# 相关知识
[JWT token](https://www.ruanyifeng.com/blog/2018/07/json_web_token-tutorial.html)