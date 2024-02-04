- [🍻 项目流程](#-项目流程)
	- [🥂 用户模块是怎么设计的?](#-用户模块是怎么设计的)
	- [视频模块设计](#视频模块设计)
	- [5. 获取已发布视频功能](#5-获取已发布视频功能)
	- [6. 拉取视频列表到首页功能](#6-拉取视频列表到首页功能)
	- [根据登录用户 ID 和查询用户 ID, 获取查询用户的详细信息](#根据登录用户-id-和查询用户-id-获取查询用户的详细信息)
	- [根据视频 ID 获取视频的点赞数量](#根据视频-id-获取视频的点赞数量)
	- [点赞视频](#点赞视频)
	- [获取关注列表](#获取关注列表)
	- [获取粉丝列表](#获取粉丝列表)
	- [关注用户](#关注用户)
- [中间件](#中间件)
	- [JWT 鉴权模块](#jwt-鉴权模块)
	- [ffmpeg 截图模块](#ffmpeg-截图模块)
	- [ftp 视频上传模块](#ftp-视频上传模块)
- [相关博客](#相关博客)

# 🍻 项目流程
## 🥂 用户模块是怎么设计的?
**需求分析**:

用户模块主要包括用户注册, 用户登录, 获取用户信息三个部分.

**相关结构**:

```go
// 用户基本信息
type TableUser struct {
	Id       		int64			// 自增 Id
	Name     		string			// 用户名
	Password 		string			// 密码
}

// 用户详细信息
type User struct {
	Id				int64			// 自增 Id
	Name			string			// 用户名
	FollowCount		int64			// 目标用户的关注数
	FollowerCount	int64			// 目标用户的粉丝数
	IsFollow		bool			// 当前用户是否关注了目标用户
	TotalFavorited	int64			// 目标用户发布视频总的被点赞量
	FavoriteCount	int64			// 目标用户点赞过多少视频
}
```

**用户注册**: 

```go
# 客户端向服务端发送注册请求
apiRouter.POST("/user/register/", controller.Register)
# 服务端从请求中获取用户名, 密码
username := c.Query("username")
password := c.Query("password")
# 服务端会先从数据库中判断该账号是否存在, 如果存在则退出
tableUser := usi.GetTableUserByUsername(username)
if username == tableUser.Name {
	c.JSON(http.StatusOK, UserLoginResponse{
		Response: Response{StatusCode: 1, StatusMsg: "User already exist"},
	})
# 如果不存在, 则创建用户基本信息对象, 并为密码进行加密存储
tableUser := dao.TableUser{
	Name:     username,
	Password: usi.EnCoder(password),
}
# 将用户信息存入数据库
usi.InsertTableUser(&tableUser)
Db.Create(&tableUser)
# 根据用户信息创建一个 token
token := usi.GenerateToken(userId, username)
# 返回响应给客户端
c.JSON(http.StatusOK, UserLoginResponse{
	Response: Response{StatusCode: 0},
	UserId:   user.Id,
	Token:    token,
})
```

**用户登录**:

```go
# 客户端向服务端发送登录请求
apiRouter.POST("/user/login/", controller.Login)
# 服务端从请求中获取用户名, 密码, 并将密码进行加密处理
username := c.Query("username")
password := c.Query("password")
encoderPassword := usi.EnCoder(password)
# 服务端从数据库中获取该账户信息, 进行比对, 如果一致则生成一个 token
tableUser = usi.GetTableUserByUsername(username)
if encoderPassword == tableUser.Password {
	token := service.GenerateToken(username)
}
# 返回响应给客户端
c.JSON(http.StatusOK, UserLoginResponse{
	Response: Response{StatusCode: 0},
	UserId:   tableUser.Id,
	Token:    token,
}) 
```

**获取用户信息**:

```go
# 客户端向服务端发送获取用户信息请求
apiRouter.GET("/user/", jwt.Auth(), controller.UserInfo)
# 服务端首先从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中, 表示当前登录用户
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 服务端从请求中获取目标用户 ID
userId := c.Query("user_id")
# 根据目标用户 ID 从数据库中获取用户的各项数据, 组装得到用户详细信息
user, err := usi.GetUserById(userId)
tableUser, err := dao.GetTableUserById(userId)
followCount, err := fsi.GetFollowingCnt(userId)			// 从 Redis / Mysql 中获取
- cnt, err := redis.RdbFollowing.SCard(redis.Ctx, strconv.Itoa(int(userId))).Result()
- redis.RdbFollowing.Expire(redis.Ctx, strconv.Itoa(int(userId)), config.ExpireTime)
- ids, err := dao.GetFollowingIds(userId)
- go addFollowingToRedis(int(userId), ids)
followerCount, err := fsi.GetFollowerCnt(userId)		// 从 Redis / Mysql 中获取
isfollow, err := fsi.IsFollowing(curId, userId)			// 从 Redis / Mysql 中获取
totalFavorited, err := lsi.TotalFavourite(userId)		// 从 Redis / Mysql 中获取
favoritedCount, err := lsi.FavouriteVideoCount(userId)	// 从 Redis / Mysql 中获取
user = User{
	Id:             userId,
	Name:           tableUser.Name,
	FollowCount:    followCount,
	FollowerCount:  followerCount,
	IsFollow:       isfollow,
	TotalFavorited: totalFavorited,
	FavoriteCount:  favoritedCount,
}
# 返回响应给客户端
c.JSON(http.StatusOK, UserResponse{
	Response: Response{StatusCode: 0},
	User:     user,
})
```
**优化设计**:

🔸 jwt token: 服务端采用 token 来识别用户身份, 其中存放着部分用户信息.

🔸 数据库安全: 数据库存储用户密码时, 存储的是 sha256 加密后的密码, 避免密码明文传输.

## 视频模块设计
**发布视频**:
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

##  5. <a name='-1'></a>获取已发布视频功能
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

##  6. <a name='-1'></a>拉取视频列表到首页功能
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

## 点赞视频
```go
// Gin 路由组监听点赞视频事件
apiRouter.POST("/favorite/action/", jwt.Auth(), controller.FavoriteAction)
// 解析 cur_id，video_id，action_type
cur_id := c.GetString("user_id")
video_id = c.Query("video_id")
action_type = c.Query("action_type")
// 进行点赞操作
err := like.FavouriteAction(cur_id, video_id, int32(action_type))
// 在 Redis 中查询有无记录
n, err := redis.RdbLikeUserId.Exists(redis.Ctx, cur_id).Result()
// 如果有，则更新缓存
_, err1 := redis.RdbLikeUserId.SAdd(redis.Ctx, cur_id, video_id).Result()
// 将数据库更新的操作放入消息队列
rabbitmq.RmqLikeAdd.Publish(sb.String())
// 如果没有，则更新缓存
_, err := redis.RdbLikeUserId.SAdd(redis.Ctx, strUserId, config.DefaultRedisValue).Result()
// 设置过期时间
_, err := redis.RdbLikeUserId.Expire(redis.Ctx, strUserId, time.Duration(config.OneMonth)*time.Second).Result()
// 根据 cur_id 获取点赞过的视频 ID 列表
videoIdList, err1 := dao.GetLikeVideoIdList(userId)
// 将 key-value 信息更新到缓存
_, err1 := redis.RdbLikeUserId.SAdd(redis.Ctx, strUserId, likeVideoId).Result()
// 将数据库更新的操作放入消息队列
```

## 获取关注列表
```go
// 通过路由组来鉴权并调用GetFollowing函数，获取当前用户的关注列表
apiRouter.GET("/relation/follow/list/", jwt.Auth(), controller.GetFollowing)
// 解析上下文中的登录的userId
userId, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
// 获取关注列表
users, err := fsi.GetFollowing(userId)
err := dao.Db.Raw(..., userId).Scan(&users)
// 输出关注列表信息
c.JSON(http.StatusOK, FollowingResp{
    UserList: users,
    Response: Response{
        StatusCode: 0,
        StatusMsg:  "OK",
    },
})
```

## 获取粉丝列表
```go
// 通过路由组来鉴权并调用GetFollowers函数，获取当前用户的粉丝列表
apiRouter.GET("/relation/follower/list", jwt.Auth(), controller.GetFollowers)
// 解析上下文中的登录的userId
userId, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
// 获取粉丝列表
users, err := fsi.GetFollowers(userId)
err := dao.Db.Raw(..., userId).Scan(&users)
// 输出粉丝列表信息
c.JSON(http.StatusOK, FollowersResp{
    Response: Response{
        StatusCode: 0,
        StatusMsg:  "OK",
    },
    UserList: users,
})
```

## 关注用户
```go
// 通过路由组来鉴权并调用RelationAction函数，实现关注功能
apiRouter.POST("/relation/action/", jwt.Auth(), controller.RelationAction)
// 获取当前用户和目标用户的ID，以及是进行关注还是取关
userId, err1 := strconv.ParseInt(c.GetString("userId"), 10, 64)
toUserId, err2 := strconv.ParseInt(c.Query("to_user_id"), 10, 64)
actionType, err3 := strconv.ParseInt(c.Query("action_type"), 10, 64)
// 进行关注
go fsi.AddFollowRelation(userId, toUserId)
// 将当前用户，目标用户ID字符串添加到消息队列，后台从中取元素写入数据库
rabbitmq.RmqFollowAdd.Publish(sb.String())
rabbitmq.InitFollowRabbitMQ()
go RmqFollowAdd.Consumer()
go f.consumerFollowAdd(msgs)
err := dao.Db.Raw(sql).Scan(nil)
// 更新 redis 中的数据
updateRedisWithAdd(userId, targetId)
// 将curId添加入targetId的粉丝列表，设置过期时间
redis.RdbFollowers.SAdd(redis.Ctx, targetIdStr, userId)
redis.RdbFollowers.Expire(redis.Ctx, targetIdStr, config.ExpireTime)
// 将targetID加入到curId的关注列表，设置过期时间
redis.RdbFollowing.SAdd(redis.Ctx, followingUserIdStr, targetId)
redis.RdbFollowing.Expire(redis.Ctx, followingUserIdStr, config.ExpireTime)
// 关注成功
c.JSON(http.StatusOK, RelationActionResp{
    Response{
        StatusCode: 0,
        StatusMsg:  "OK",
    },
})
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

## ffmpeg 截图模块
**结构体**
```go
// 存放视频名和封面名
type Ffmsg struct {
	VideoName string
	ImageName string
}
```
**Init()**
```go
// 创建一个 ssh 连接对象，连接到服务器
ClientSSH, err = ssh.Dial("tcp", addr, SSHconfig)
// 创建一个管道，用于存放视频名，封面名信息
Ffchan = make(chan Ffmsg, config.MaxMsgCount)
// 调用协程，从管道中取出一个数据进行处理，并且保持长连接
go dispatcher()
go keepAlive()
// dispatcher 处理，就是循环从管道取数据，截取封面
for ffmsg := range Ffchan {
    go func(f Ffmsg) { err := Ffmpeg(f.VideoName, f.ImageName) }(ffmsg)
}
// 通过远程调用ffmpeg命令来截图, 截取的图放在服务器中的指定路径
session, err := ClientSSH.NewSession()
combo, err := session.CombinedOutput("ls;/usr/.../ffmpeg -ss 00:00:01 -i /video_path/" + videoName + ".mp4 -vframes 1 /images_path/" + imageName + ".jpg")
```

## ftp 视频上传模块
**InitFTP()**
```go
// 初始化一个ftp连接对象
MyFTP, err = goftp.Connect(config.ConConfig)
// 登录上tcp服务器
err = MyFTP.Login(config.FtpUser, config.FtpPsw)
// 登录成功后维持长连接
go keepAlive()
```

# 相关博客
[JWT token](https://www.ruanyifeng.com/blog/2018/07/json_web_token-tutorial.html)
[Nginx](https://juejin.cn/post/6844904129987526663)
[正向代理 / 反向代理](https://juejin.cn/post/6844904129987526663)