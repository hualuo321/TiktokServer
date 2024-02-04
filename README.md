- [🍺 项目流程](#-项目流程)
	- [🍻 用户模块是怎么设计的?](#-用户模块是怎么设计的)
	- [🍻 视频模块是怎么设计的?](#-视频模块是怎么设计的)
	- [🍻 点赞模块是怎么设计的?](#-点赞模块是怎么设计的)
	- [🍻 评论模块是怎么设计的?](#-评论模块是怎么设计的)
	- [🍻 关注模块是怎么设计的?](#-关注模块是怎么设计的)
- [🍺 相关知识](#-相关知识)
	- [🍻 jwt token 是什么?](#-jwt-token-是什么)
	- [🍻 Nginx 正向代理和反向代理?](#-nginx-正向代理和反向代理)
- [🍺 项目遇到的问题](#-项目遇到的问题)
	- [🍻 Redis 中的脏读现象?](#-redis-中的脏读现象)
- [🍺 相关博客](#-相关博客)

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

## 🍻 评论模块是怎么设计的?

**需求分析**:

评论模块主要包括发布评论, 删除评论, 查看评论三个部分.

**相关结构**

```go
// 评论基本信息
type TableComment struct {
	Id          int64     // 评论id
	UserId      int64     // 评论用户id
	VideoId     int64     // 视频id
	CommentText string    // 评论内容
	CreateDate  time.Time // 评论发布的日期
	Cancel      int32     // 取消评论为1，发布评论为0
}

// 评论扩展信息
type Comment struct {
	Id         	int64	// 评论 ID
	UserInfo   User		// 发布评论的用户
	Content    string	// 评论内容
	CreateDate string	// 发布日期
}
```

**发布评论**:

```go
# 客户端向服务端发送发布评论请求
apiRouter.POST("/comment/action/", jwt.Auth(), controller.CommentAction)
# 服务端首先从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 从请求中获取当前用户 ID, 视频 ID, 发布动作, 评论内容, 并对垃圾评论进行过滤
curId, _ := strconv.ParseInt(c.GetString("curId"), 10, 64)
videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
actionType, err := strconv.ParseInt(c.Query("action_type"), 10, 32)
content := c.Query("comment_text")
content = util.Filter.Replace(content, '#')
# 根据上述信息发布评论
comment, err := csi.Send(tableComment)
# 存储评论信息到数据库中
tableComment, err := dao.InsertComment(tableComment)
Db.Model(TableComment{}).Create(&tableComment)
# 获取当前用户的结构体, 拼接到评论详细信息中
user, err := usi.GetUserByIdWithCurId(tableComment.UserId, tableComment.UserId)
var comemnt = Comment{}
comment.userInfo = user
# 将评论信息更新到 Redis 中
insertRedisVideoCommentId(strconv.Itoa(int(comment.VideoId)), strconv.Itoa(int(commentRtn.Id)))
redis.RdbVCid.SAdd(redis.Ctx, videoId, commentId).Result()
redis.RdbCVid.Set(redis.Ctx, commentId, videoId, 0).Result()
# 返回响应给客户端
c.JSON(http.StatusOK, CommentActionResponse{
	StatusCode: 0,
	StatusMsg:  "send comment success",
	Comment:    commentInfo,
})
```

**删除评论**:

```go
# 客户端向服务端发送发布评论请求
apiRouter.POST("/comment/action/", jwt.Auth(), controller.CommentAction)
# 服务端首先从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 从请求中获取当前用户 ID, 视频 ID, 发布动作, 评论 ID
curId, _ := strconv.ParseInt(c.GetString("curId"), 10, 64)
videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
actionType, err := strconv.ParseInt(c.Query("action_type"), 10, 32)
commentId, err := strconv.ParseInt(c.Query("comment_id"), 10, 64)
# 根据上述信息删除评论
commentService.DelComment(commentId)
# 先检查 Redis 中是否存在记录, 如果有则删除, 消息队列更新数据库
redis.RdbCVid.Exists(redis.Ctx, strconv.FormatInt(commentId, 10)).Result()
redis.RdbCVid.Get(redis.Ctx, strconv.FormatInt(commentId, 10)).Result()
redis.RdbCVid.Del(redis.Ctx, strconv.FormatInt(commentId, 10)).Result()
redis.RdbVCid.SRem(redis.Ctx, vid, strconv.FormatInt(commentId, 10)).Result()
rabbitmq.RmqCommentDel.Publish(strconv.FormatInt(commentId, 10))
# 如果 Redis 中不存在记录, 则直接删除数据库中数据
dao.DeleteComment(commentId)
Db.Model(Comment{}).Where(map[string]interface{}{"id": id, "cancel": config.ValidComment}).First(&commentInfo)
Db.Model(Comment{}).Where("id = ?", id).Update("cancel", config.InvalidComment)
# 返回响应给客户端
c.JSON(http.StatusOK, CommentActionResponse{
	StatusCode: 0,
	StatusMsg:  "delete comment success",
})
```

**获取评论列表**:

```go
# 客户端向服务端发送获取评论列表请求
apiRouter.GET("/comment/list/", jwt.AuthWithoutLogin(), controller.CommentList)
# 服务端首先从请求中获取 token 进行解析, 不论有无携带 token, 都能进行获取评论列表功能
auth := context.Query("token")
if len(auth) == 0 {curId = "0"} break
token, err := parseToken(auth)
curId = token.Id
# 从请求中获取用户 ID, 视频 ID 等信息
curId, _ := strconv.ParseInt(c.GetString("curId"), 10, 64)
videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
# 根据上述信息获取视频的评论列表
commentList, err := csi.GetList(videoId, curId)
# 首先从 Redis 中获取数据, 如果存在记录则获取
redis.RdbVCid.SCard(redis.Ctx, strconv.FormatInt(videoId, 10)).Result()
# 如果 Redis 中不存在记录, 则从数据库中获取评论基本信息列表, 转变为评论扩展信息列表
commentList, err := dao.GetCommentList(videoId)
oneComment(&commentData, &comment, userId)
# 并更新到 Redis 中
redis.RdbVCid.SAdd(redis.Ctx, strconv.Itoa(int(videoId)), config.DefaultRedisValue).Result()
redis.RdbVCid.Expire(redis.Ctx, strconv.Itoa(int(videoId)), time.Duration(config.OneMonth)*time.Second).Result()
insertRedisVideoCommentId(strconv.Itoa(int(videoId)), strconv.Itoa(int(_comment.Id)))
redis.RdbVCid.SAdd(redis.Ctx, videoId, commentId).Result()
redis.RdbCVid.Set(redis.Ctx, commentId, videoId, 0).Result()
```

**优化设计**:

🔸 当服务器直接与 Mysql 进行交互时, 客户端的响应时间较慢, 为了减少响应时间而使用了具有高性能的 Redis 缓存. 当用户在刷评论时, 最常用到的是获取评论列表功能, 当用户进行相关操作时, 直接从 Redis 中获取数据进行响应，提高用户操作的流畅度.

🔸 当大量用户同时向服务器发出请求时, 如果直接对数据库进行处理, 那么数据库压力过大可能会导致宕机. 因此在项目中采用 rabbitMQ 作为消息队列, 当需要对数据库进行操作时, 将操作放入消息队列中, 由服务器从消息队列中取消息, 不断地进行处理.

🔸 当用户获取视频的评论列表时, 查询的都是当前视频的评论, 为了优化查询的性能, 将视频 ID 作为评论表的索引, 增加查询速度.

```sql
CREATE INDEX idx_video_id ON comment(video_id);
```

🔸 当对视频扩展信息进行封装时, 需要获取当前视频的评论量, 如果直接从数据库里查询会很慢, 但是采用 Redis 可以直接获取视频 key 对应 value 的长度大小作为评论的数量, 速度很快.

## 🍻 关注模块是怎么设计的?

**需求分析**:

关注模块主要包括关注操作, 取关操作, 获取关注列表, 获取粉丝列表四个部分.

**相关结构**:

```go
# 关注信息
type Follow struct {
	Id         	int64 	// 自增 ID
	UserId     	int64	// 发起关注方
	FollowerId 	int64	// 被关注方
	Cancel     	int8	// 是否关注
}
```

**关注操作**

```go
# 客户端向服务端发送关注请求
apiRouter.POST("/relation/action/", jwt.Auth(), controller.RelationAction)
# 服务端从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 服务端从请求中获取当前用户 ID, 目标用户 ID, 关注类型
curId, err1 := strconv.ParseInt(c.GetString("curId"), 10, 64)
userId, err2 := strconv.ParseInt(c.Query("userId"), 10, 64)
actionType, err3 := strconv.ParseInt(c.Query("action_type"), 10, 64)
# 服务端根据上述参数进行关注操作
go fsi.AddFollowRelation(curId, userId)
# 接着更新 Redis 缓存中的内容
updateRedisWithAdd(userId, targetId)
redis.RdbFollowers.SCard(redis.Ctx, userId).Result()
redis.RdbFollowers.SAdd(redis.Ctx, userId, curId)
redis.RdbFollowers.Expire(redis.Ctx, userId, config.ExpireTime)
redis.RdbFollowing.SCard(redis.Ctx, curId).Result()
redis.RdbFollowing.SAdd(redis.Ctx, curId, userId)
redis.RdbFollowing.Expire(redis.Ctx, curId, config.ExpireTime)
# 将更新数据库操作写入消息队列
rabbitmq.RmqFollowAdd.Publish(sb.String())
# 服务端向客户端返回结果
c.JSON(http.StatusOK, RelationActionResp{
	Response{
		StatusCode: 0,
		StatusMsg:  "OK",
	},
}) 
```

**取关操作**:

```go
# 客户端向服务端发送取关请求
apiRouter.POST("/relation/action/", jwt.Auth(), controller.RelationAction)
# 服务端从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 服务端从请求中获取当前用户 ID, 目标用户 ID, 关注类型
curId, err1 := strconv.ParseInt(c.GetString("curId"), 10, 64)
userId, err2 := strconv.ParseInt(c.Query("userId"), 10, 64)
actionType, err3 := strconv.ParseInt(c.Query("action_type"), 10, 64)
# 服务端根据上述参数进行关注操作
go fsi.DeleteFollowRelation(curId, uerId)
# 接着更新 Redis 缓存中的内容
updateRedisWithAdd(userId, targetId)
redis.RdbFollowers.SCard(redis.Ctx, userId).Result()
redis.RdbFollowers.SRem(redis.Ctx, userId, curId)
redis.RdbFollowers.Expire(redis.Ctx, userId, config.ExpireTime)
redis.RdbFollowing.SCard(redis.Ctx, curId).Result()
redis.RdbFollowing.SRem(redis.Ctx, curId, userId)
redis.RdbFollowing.Expire(redis.Ctx, curId, config.ExpireTime)
# 将更新数据库操作写入消息队列
rabbitmq.RmqFollowDel.Publish(sb.String())
# 服务端向客户端返回结果
c.JSON(http.StatusOK, RelationActionResp{
	Response{
		StatusCode: 0,
		StatusMsg:  "OK",
	},
}) 
```

**获取关注列表**:

```go
# 客户端向服务端发送获取关注列表请求
apiRouter.GET("/relation/follow/list/", jwt.Auth(), controller.GetFollowing)
# 服务端从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 服务端从请求中获取目标用户 ID
userId, err2 := strconv.ParseInt(c.Query("userId"), 10, 64)
# 利用上述参数尝试获取目标用户所关注的用户
userList = fsi.GetFollowing(userId)
# 首先会查看缓存中是否有记录, 如果有则获取数据
redis.RdbFollowers.SCard(redis.Ctx, followingIdStr).Result()
userIdList = redis.RdbFollowing.SMembers(redis.Ctx, userId).Result()
# 根据这些 userIdList 来利用协程并发获取数据, 转化为 userList
userList := make([]User, len)
go usi.GetUserByIdWithCurId(userList[i], userId)
- followerCount, err := fsi.GetFollowerCnt(userId)		// 从 Redis / Mysql 中获取
- isfollow, err := fsi.IsFollowing(curId, userId)			// 从 Redis / Mysql 中获取
- totalFavorited, err := lsi.TotalFavourite(userId)		// 从 Redis / Mysql 中获取
- favoritedCount, err := lsi.FavouriteVideoCount(userId)	// 从 Redis / Mysql 中获取
# 如果缓存中没有, 则从数据库中获取数据, 更新到缓存中
userIdList = dao.GetFollowingIds(userId)
Db.Model(Follow{}).Where("follower_id = ?", userId).Pluck("user_id", &ids)
go setRedisFollowing(userId, userList)
redis.RdbFollowing.SAdd(redis.Ctx, userId, DefaultRedisValue)
redis.RdbFollowing.Expire(redis.Ctx, followingIdStr, config.ExpireTime)
redis.RdbFollowing.SAdd(redis.Ctx, userId, userList[i].id)
# 服务器向客户端返回数据
c.JSON(http.StatusOK, FollowingResp{
	UserList: userList,
	Response: Response{
		StatusCode: 0,
		StatusMsg:  "OK",
	},
})
```

**获取粉丝列表**:

```go
# 客户端向服务端发送获取关注列表请求
apiRouter.GET("/relation/follower/list", jwt.Auth(), controller.GetFollowers)
# 服务端从请求中获取 token 进行解析, 如果解析正确, 则将 token 中的用户信息添加到上下文中
auth := context.Query("token")
token, err := parseToken(auth)
context.Set("curId", token.Id)
context.Next()
# 服务端从请求中获取目标用户 ID
userId, err2 := strconv.ParseInt(c.Query("userId"), 10, 64)
# 利用上述参数尝试获取目标用户所关注的用户
userList = fsi.GetFollowers(userId)
# 首先会查看缓存中是否有记录, 如果有则获取数据
redis.RdbFollowers.SCard(redis.Ctx, followersIdStr).Result()
userIdList = redis.RdbFollowers.SMembers(redis.Ctx, userId).Result()
# 根据这些 userIdList 来利用协程并发获取数据, 转化为 userList
userList := make([]User, len)
go usi.GetUserByIdWithCurId(userList[i], userId)
- followerCount, err := fsi.GetFollowerCnt(userId)		// 从 Redis / Mysql 中获取
- isfollow, err := fsi.IsFollowing(curId, userId)		// 从 Redis / Mysql 中获取
- totalFavorited, err := lsi.TotalFavourite(userId)		// 从 Redis / Mysql 中获取
- favoritedCount, err := lsi.FavouriteVideoCount(userId)// 从 Redis / Mysql 中获取
# 如果缓存中没有, 则从数据库中获取数据, 更新到缓存中
userIdList = dao.GetFollowersIds(userId)
Db.Model(Follow{}).Where("user_id = ?", userId).Where("cancel = ?", 0).Pluck("follower_id", &ids)
go setRedisFollowers(userId, userList)
redis.RdbFollowers.SAdd(redis.Ctx, userId, DefaultRedisValue)
redis.RdbFollowers.Expire(redis.Ctx, userId, config.ExpireTime)
redis.RdbFollowers.SAdd(redis.Ctx, userId, userList[i])
# 服务器向客户端返回数据
c.JSON(http.StatusOK, FollowersResp{
	Response: Response{
		StatusCode: 0,
		StatusMsg:  "OK",
	},
	UserList: users,
})
```

**优化设计**:

🔸 当服务器直接与 Mysql 进行交互时, 客户端的响应时间较慢, 为了减少响应时间而使用了具有高性能的 Redis 缓存. 当用户在获取关注列表时, 直接从 Redis 中获取数据进行响应，提高用户操作的流畅度.
🔸 当大量用户同时向服务器发出请求时, 如果直接对数据库进行处理, 那么数据库压力过大可能会导致宕机. 因此在项目中采用 rabbitMQ 作为消息队列, 当需要对数据库进行操作时, 将操作放入消息队列中, 由服务器从消息队列中取消息, 不断地进行处理.
🔸 考虑到关注取关操作时, 会先判断用户双方是否关注过, 涉及到当前用户 ID 和目标用户 ID, 所以可以采用复合索引来提升搜索的速度.

```sql
CREATE INDEX cur_id_to_target_id_idx ON follows(cur_id, target_id) USING BTREE;
```

# 🍺 相关知识
## 🍻 jwt token 是什么?

🔸 jwt token 是一种跨域认证的解决方案.

🔸 互联网中常用的用户认证是通过 session 和 cookie 实现的. 当客户端向服务器发送用户和密码时, 服务端验证通过后会会在当前会话中存放用户信息, 并在响应中返回一个 session_id 放在 cookie 中. 当客户端之后再进行请求时, 服务端可以通过 cookie 中的 session_id 来识别客户端的身份.

🔸 然而这种方式在服务器集群中, 需要要求 session 数据共享, 每台服务器都能获取 session. 另一方面, 当用户量很大时存储 session 的内存占用也很多. token 实际上就是将用户信息保存到客户端, 每次请求的时候都携带 token 发送到服务端, 服务端只验证 token 是否有效. Get 请求中, jwt 可以放在 url 里, Post 请求中, jwt 可以放在请求体里.

🔸 jwt 主要分为 3 个部分, 头部, 荷载, 签名. 其中头部存放 jwt 元数据, 比如令牌类型, 荷载存放需要传递的数据, 比如用户信息, 签名是对前两个部分的签名, 防止数据被篡改.

## 🍻 Nginx 正向代理和反向代理?

🔸 正向代理: 是指客户端向服务器发送请求时, 会通过一个代理服务器间接访问服务器, 代理服务器会转交客户端的请求, 从服务器获取内容并返回给客户端. 其中服务器并不知道请求客户端的具体地址..

🔸 反向代理: 是指客户端向服务器发送请求时, 并不知道该服务器是代理服务器, 代理服务器会将客户端的请求进行转发, 从其他服务器上获取内容并返回给客户端. 其中客户端并不清楚访问服务器的具体地址.

# 🍺 项目遇到的问题
## 🍻 Redis 中的脏读现象?

**脏读介绍**:

🔸 脏读: 通常是指事务 A 读取到了事务 B 修改但未提交的数据, 如果事务 B 发生回滚, 那么事务 A 读取的数据和数据库中的数据会不一致.

**存在问题**:

🔸 在项目中的视频点赞模块, 如果当前用户取消对最后一个视频的点赞, 那么 Redis 会将这个用户的 key 删除.

🔸 在同一时间上, 如果用户又对另一个视频点赞, 那么 Redis 发现 key 不存在时, 就会从数据库中查找数据, 为用户更新 key.

🔸 由于网络延迟等原因, 取消点赞还没更新到数据库, 那么其他用户查询这个 key 时, 获取到的点赞列表和实际列表不一致, 出现脏读.

**解决方案**: 

🔸 可以在初始化时, 为 key 添加一个默认值, 并设置一个 key 过期时间, 那么就算之后点赞列表为空, key 也不会被删除, 只会通过过期策略来删除.

🔸 当进行点赞/取消等操作时, 会先对 Redis 中的数据进行更新. 数据库中的数据通过消息队列来更新. 

🔸 当其他用户查询这个 key 时, 会直接从 Redis 里查询, 避免了从数据库更新过慢导致的脏读现象.


# 🍺 相关博客
[JWT token](https://www.ruanyifeng.com/blog/2018/07/json_web_token-tutorial.html)

[Nginx](https://juejin.cn/post/6844904129987526663)

[正向代理 / 反向代理](https://juejin.cn/post/6844904129987526663)