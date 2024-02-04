- [🍺 项目流程](#-项目流程)
	- [🍻 用户模块是怎么设计的?](#-用户模块是怎么设计的)
	- [🍻 视频模块是怎么设计的?](#-视频模块是怎么设计的)
	- [🍻 点赞模块是怎么设计的?](#-点赞模块是怎么设计的)
- [中间件](#中间件)
	- [JWT 鉴权模块](#jwt-鉴权模块)
	- [ffmpeg 截图模块](#ffmpeg-截图模块)
	- [ftp 视频上传模块](#ftp-视频上传模块)
- [相关博客](#相关博客)

# 🍺 项目流程
## 🍻 用户模块是怎么设计的?
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

🔸 jwt token: 服务端采用 token 来识别用户身份, 其中存放着部分用户信息. 同时设置了不同的权限, 比如发布视频则必须要携带正确的 token, 确保用户登录. 而刷新 Feed 视频流则不需要强制携带 token, 非登录状态也能刷视频.

🔸 数据库安全: 数据库存储用户密码时, 存储的是 sha256 加密后的密码, 避免密码明文传输.

## 🍻 视频模块是怎么设计的?

**需求分析**:

视频模块主要包括发布视频, 获取视频发布列表, 获取视频 Feed 流三个部分.

**相关结构**

```go
# 视频基本信息
type TableVideo struct {
	Id				int64		// 自增 Id
	AuthorId		int64		// 作者 Id
	PlayUrl			string		// 视频地址
	CoverUrl		string		// 封面地址
	PublishTime 	time.Time	// 发布时间
	Title			string		// 视频标题
}

# 视频详细信息
type Video struct {
	dao.TableVideo				// 视频基本信息
	Author        	User		// 视频作者
	FavoriteCount 	int64		// 视频被点赞量
	CommentCount  	int64		// 视频的评论数
	IsFavorite    	bool		// 当前用户是否点赞了该视频
}
```

**发布视频**:

```go
# 客户端向服务端发送发布视频请求
apiRouter.POST("/publish/action/", jwt.AuthBody(), controller.Publish)
# 服务端首先从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 服务器从请求中获取目标用户 ID, 发布视频的数据, 发布视频的标题
userId, _ := strconv.ParseInt(c.GetString("userId"), 10, 64)
data, err := c.FormFile("data")
title := c.PostForm("title")
# 根据获取到的上述信息, 发布视频到 FTP 服务器
err = vsi.Publish(data, userId, title)
err = dao.VideoFTP(file, videoName)
err = ftp.MyFTP.Stor(videoName+".mp4", file)
# 在 FTP 服务器上执行 ffmpeg 命令来远程对视频截图作为封面, 同样保存在 TFP 服务器中
imageName := uuid.NewV4().String()
session, err := ClientSSH.NewSession()
session.CombinedOutput("ls;/ffmpeg/path/ -ss 00:00:01 -i /video/path/" + videoName + ".mp4 -vframes 1 /images/path/" + imageName + ".jpg")
# 将基本视频信息保存在数据库中
err = dao.Save(videoName, imageName, userId, title)
var video TableVideo {video.PublishTime = time.Now(), ...}
Db.Save(&video)
# 返回响应给客户端
c.JSON(http.StatusOK, Response{
	StatusCode: 0,
	StatusMsg:  "uploaded successfully",
})
```

**获取视频列表**:

```go
# 客户端向服务端发送获取视频列表请求
apiRouter.GET("/publish/list/", jwt.Auth(), controller.PublishList)
# 服务端首先从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 服务器从请求中获取目标用户 ID, 当前用户 ID
userId, _ := strconv.ParseInt(c.GetQuery("userId"), 10, 64)
curId, _ := strconv.ParseInt(c.GetString("curId"), 10, 64)
# 根据目标用户 ID 获取他的发布视频列表
videoList, err := vsi.List(userId, curId)
# 首先会从数据库中进行查询, 获取视频基本信息列表
tableVideoList, err := dao.GetTableVideoByUserId(userId)
Db.Where(&TableVideo{AuthorId: userId}).Find(&tableVideoList)
# 根据视频基本信息来组装成为视频详细信息, 调用携程并发写入
videoService.creatVideo(&video, &temp, userId)
wg.Add(4)
go func() {video.Author, err = vsi.GetUserByIdWithCurId(data.AuthorId, userId), wg.Done()}()
go func() {video.FavoriteCount, err = vsi.FavouriteCount(data.Id), wg.Done()}()
go func() {video.CommentCount, err = vsi.CountFromVideoId(data.Id), wg.Done()}()
go func() {video.IsFavorite, err = vsi.IsFavourite(video.Id, userId)), wg.Done()}()
wg.Wait()
# 返回响应给客户端
c.JSON(http.StatusOK, VideoListResponse{
	Response:  Response{StatusCode: 0},
	VideoList: videoList,
})
```

**获取视频 Feed 流**

```go
# 客户端向服务端发送获取视频 Feed 流请求
apiRouter.GET("/feed/", jwt.AuthWithoutLogin(), controller.Feed)
# 服务端首先从请求中获取 token 进行解析, 不论有无携带 token, 都能进行刷新视频首页功能
auth := context.Query("token")
if len(auth) == 0 {curId = "0"} break
token, err := parseToken(auth)
curId = token.Id
# 服务器从请求中获取请求时间, 当前用户 ID
inputTime := c.Query("latest_time")
curId, _ := strconv.ParseInt(c.GetString("curId"), 10, 64)
# 根据上述信息获取最新的视频详细信息列表
feed, nextTime, err := videoService.Feed(lastTime, userId)
# 首先会从数据库中进行查询, 根据请求时间获取在其之前视频基本信息列表
tableVideos, err := dao.GetTableVideosByLastTime(lastTime)
Db.Where("publish_time<?", lastTime).Order("publish_time desc").Limit(config.VideoCount).Find(&videos)
# 根据视频基本信息来组装成为视频详细信息, 调用携程并发写入
videoService.creatVideo(&video, &temp, userId)
wg.Add(4)
go func() {video.Author, err = vsi.GetUserByIdWithCurId(data.AuthorId, userId), wg.Done()}()
go func() {video.FavoriteCount, err = vsi.FavouriteCount(data.Id), wg.Done()}()
go func() {video.CommentCount, err = vsi.CountFromVideoId(data.Id), wg.Done()}()
go func() {video.IsFavorite, err = vsi.IsFavourite(video.Id, userId)), wg.Done()}()
wg.Wait()
# 返回响应给客户端
c.JSON(http.StatusOK, FeedResponse{
	Response:  Response{StatusCode: 0},
	VideoList: feed,
	NextTime:  nextTime.Unix(),	// 视频详细信息列表中的最早发布时间
})
```

**优化设计**:

🔸 在获取视频 Feed 流和获取发布视频列表时, 首先会从数据库中获取视频基本信息, 根据视频基本信息调用其他服务组装获取视频详细信息, 大量的同步调用会使得调用缓慢, 影响用户的体验. 在项目通过 go 携程并行调用其他服务来缩短信息拼装的整体时间. 

🔸 在根据视频基本信息调用其他服务组装视频详细信息时, 通过引用的方式将基本信息对象嵌入到详细信息对象中, 避免了资源的拷贝操作.

🔸 在视频发布功能中, 原本是在服务器进行截图, 然后通过建立两个 ftp 连接, 将视频和封面数据都上传到 FTP 服务器, 但是这样数据传输的流量会更大. 且需要两个 ftp 连接. 在项目中仅建立一个 ftp 连接传输视频数据, 通过 ssh 连接 FTP 服务器远程调用 ffmpeg 命令截图, 在 FTP 服务器上获取封面数据.

🔸 在连接中, 将 ssh 和 ftp 连接均设置为长连接, 减少连接断开的情况发生.

## 🍻 点赞模块是怎么设计的?

**需求分析**:

点赞模块包括点赞, 取消点赞, 获取点赞列表三个部分.

**相关结构**:

```go
// 点赞基本信息
type Like struct {
	Id      	int64 	// 自增 Id
	UserId  	int64 	// 点赞方
	VideoId 	int64 	// 被点赞视频
	Cancel  	int8  	// 是否点赞，0为点赞，1为取消赞
}
```

**点赞**:

```go
# 客户端向服务端发送点赞请求
apiRouter.POST("/favorite/action/", jwt.Auth(), controller.FavoriteAction)
# 服务端首先从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 从请求中获取当前用户 ID, 被点赞视频 ID, 点赞类型
curId, _ := strconv.ParseInt(c.GetString("curId"), 10, 64)
videoId, _ := strconv.ParseInt(c.Query("video_id"), 10, 64)
actionType, _ := strconv.ParseInt(c.Query("action_type"), 10, 64)
# 调用点赞动作
lsi.FavouriteAction(curId, videoId, int32(actionType))
# 先从 Redis 中查找有无当前用户的 key, 如果有则添加一个 videoId, 并将更新数据库操作放入消息队列中
redis.RdbLikeUserId.Exists(redis.Ctx, curId).Result()
redis.RdbLikeUserId.SAdd(redis.Ctx, curId, videoId).Result()
rabbitmq.RmqLikeAdd.Publish(sb.String())
# 如果 Redis 中找不到, 则新建一个 key, 设置过期时间, 并添加初始值防止脏读
redis.RdbLikeUserId.SAdd(redis.Ctx, curId, config.DefaultRedisValue).Result()
redis.RdbLikeUserId.Expire(redis.Ctx, curId, time.Duration(config.OneMonth)*time.Second).Result()
# 从数据库中读取点赞信息列表, 更新到 Redis 中, 在通过消息队列更新数据库
videoIdList, err1 := dao.GetLikeVideoIdList(curId)
for likeVideoId := range videoIdList {
	redis.RdbLikeUserId.SAdd(redis.Ctx, userId, likeVideoId).Result(); err1 != nil
}
redis.RdbLikeUserId.SAdd(redis.Ctx, curId, videoId).Result()
rabbitmq.RmqLikeAdd.Publish(sb.String())
# 同时更新视频被那些用户点赞的 Redis 和 Mysql 数据, 流程和上述一样
# 不过这里不用消息队列, 因为 Redis 中的 RdbLikeUserId 和 RdbLikeVideoId 对应的是一张 Like 表
redis.RdbLikeVideoId.Exists(redis.Ctx, VideoId).Result()
redis.RdbLikeVideoId.SAdd(redis.Ctx, VideoId, curId).Result()
...
# 返回响应给客户端
c.JSON(http.StatusOK, likeResponse{
	StatusCode: 0,
	StatusMsg:  "favourite action success",
})
```

**取消点赞**:

```go
# 客户端向服务端发送取消点赞请求
apiRouter.POST("/favorite/action/", jwt.Auth(), controller.FavoriteAction)
# 服务端首先从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 从请求中获取当前用户 ID, 被点赞视频 ID, 点赞类型
curId, _ := strconv.ParseInt(c.GetString("curId"), 10, 64)
videoId, _ := strconv.ParseInt(c.Query("video_id"), 10, 64)
actionType, _ := strconv.ParseInt(c.Query("action_type"), 10, 64)
# 调用点赞动作
lsi.FavouriteAction(curId, videoId, int32(actionType))
# 先从 Redis 中查找有无当前用户的 key, 如果有则删除 videoId, 并将更新数据库操作放入消息队列中
redis.RdbLikeUserId.Exists(redis.Ctx, curId).Result()
redis.RdbLikeUserId.SRem(redis.Ctx, curId, videoId).Result()
rabbitmq.RmqLikeDel.Publish(sb.String())
# 如果 Redis 中找不到, 则新建一个 key, 设置过期时间, 并添加初始值防止脏读
redis.RdbLikeUserId.SAdd(redis.Ctx, curId, config.DefaultRedisValue).Result()
redis.RdbLikeUserId.Expire(redis.Ctx, curId, time.Duration(config.OneMonth)*time.Second).Result()
# 从数据库中读取点赞信息列表, 更新到 Redis 中, 在通过消息队列更新数据库
videoIdList, err1 := dao.GetLikeVideoIdList(curId)
for likeVideoId := range videoIdList {
	redis.RdbLikeUserId.SAdd(redis.Ctx, userId, likeVideoId).Result(); err1 != nil
}
redis.RdbLikeUserId.SRem(redis.Ctx, curId, videoId).Result()
redis.RdbLikeUserId.Del(redis.Ctx, curId)
# 同时更新视频被那些用户点赞的 Redis 和 Mysql 数据, 流程和上述一样
# 不过这里不用消息队列, 因为 Redis 中的 RdbLikeUserId 和 RdbLikeVideoId 对应的是一张 Like 表
redis.RdbLikeVideoId.Exists(redis.Ctx, VideoId).Result()
redis.RdbLikeVideoId.SRem(redis.Ctx, VideoId, curId).Result()
...
# 返回响应给客户端
c.JSON(http.StatusOK, likeResponse{
	StatusCode: 0,
	StatusMsg:  "favourite action success",
})
```

**获取点赞列表**:

```go
# 客户端向服务端发送获取点赞列表请求
apiRouter.GET("/favorite/list/", jwt.Auth(), controller.GetFavouriteList)
# 服务端首先从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 从请求中获取当前用户 ID, 目标用户 ID
curId, _ := strconv.ParseInt(c.GetString("curId"), 10, 64)
userId, _ := strconv.ParseInt(c.GetQuery("userId"), 10, 64)
# 通过目标用户 ID 获取他点赞过的视频列表
videos, err := lsi.GetFavouriteList(userId, curId)
# 首先从 Redis 中查询有无 key 为 userId, 如果有则获取相应的 videoIdList
redis.RdbLikeUserId.Exists(redis.Ctx, strUserId).Result()
videoIdList, err1 := redis.RdbLikeUserId.SMembers(redis.Ctx, strUserId).Result()
# 通过 videoId 来获取视频基本信息, 再通过基本信息来获取视频详细信息
go lsi.addFavouriteVideoList(videoId, curId, favoriteVideoList, &wg)
tableVideo, err := like.GetVideo(videoId, curId)
video, err := dao.GetVideoByVideoId(videoId)
videoService.creatVideo(&video, &temp, userId)
wg.Add(4)
go func() {video.Author, err = vsi.GetUserByIdWithCurId(data.AuthorId, userId), wg.Done()}()
go func() {video.FavoriteCount, err = vsi.FavouriteCount(data.Id), wg.Done()}()
go func() {video.CommentCount, err = vsi.CountFromVideoId(data.Id), wg.Done()}()
go func() {video.IsFavorite, err = vsi.IsFavourite(video.Id, userId)), wg.Done()}()
wg.Wait()
# 如果 Redis 中没有相关记录, 则从 Mysql 中查询并更新到 Redis
redis.RdbLikeUserId.SAdd(redis.Ctx, strUserId, config.DefaultRedisValue).Result()
redis.RdbLikeUserId.Expire(redis.Ctx, strUserId, time.Duration(config.OneMonth)*time.Second).Result()
videoIdList, err1 := dao.GetLikeVideoIdList(userId)
redis.RdbLikeUserId.SAdd(redis.Ctx, strUserId, likeVideoId).Result()
# 通过 videoId 来获取视频基本信息, 再通过基本信息来获取视频详细信息
go lsi.addFavouriteVideoList(videoId, curId, favoriteVideoList, &wg)
...
# 返回响应给客户端
c.JSON(http.StatusOK, GetFavouriteListResponse{
	StatusCode: 0,
	StatusMsg:  "get favouriteList success",
	VideoList:  videos,
})
```

**优化设计**：

🔸 当服务器直接与 Mysql 进行交互时, 客户端的响应时间较慢, 为了减少响应时间而使用了具有高性能的 Redis 缓存. 当用户在刷视频时, 最常用到的功能是点赞, 取消赞功能, 当用户进行相关操作时, 直接从 Redis 中获取数据进行响应 ，提高用户操作的流畅度.

🔸 当大量用户同时向服务器发出请求时, 如果直接对数据库进行处理, 那么数据库压力过大可能会导致宕机. 因此在项目中采用 rabbitMQ 作为消息队列, 当需要对数据库进行操作时, 将操作放入消息队列中, 由服务器从消息队列中取消息, 不断地进行处理.

🔸 在 Redis 中 key 的初始化时, 会为 key 添加一个默认值, 并设置一个过期时间, 那么就算之后点赞列表为空, key 也不会被删除, 只会通过过期策略来删除. 当进行点赞 / 取消点赞等操作时, 会先对 Redis 中的数据进行更新, 数据库中的数据通过消息队列来更新. 当其他用户查询这个 key 时, 会直接从 Redis 里查询, 避免了从数据库更新过慢导致的脏读现象.

🔸 当获取点赞视频列表时, 最初是向 Mysql 中查找符合条件的 videoID, 再获取获取一个完整的 video 对象, 涉及到多张表的查询, 响应速度很慢. 现在是从 Redis 中获取符合条件的 videoId, 再通过协程的方式并发获取 video 信息, 提高了响应速度.



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