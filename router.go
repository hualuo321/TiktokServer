package main

import (
	"TikTok/controller"
	"TikTok/middleware/jwt"
	"github.com/gin-gonic/gin"
)

func initRouter(r *gin.Engine) {
	apiRouter := r.Group("/douyin")
	// 基础 API
	// 1.1 Gin 路由组监听用户登录事件
	apiRouter.POST("/user/login/", controller.Login)
	// 2.1 Gin 路由组监听用户注册事件
	apiRouter.POST("/user/register/", controller.Register)
	// 3.1 Gin 路由组监听获取用户信息事件
	apiRouter.GET("/user/", jwt.Auth(), controller.UserInfo)
	// 4.1 Gin 路由组监听发布视频事件
	apiRouter.POST("/publish/action/", jwt.AuthBody(), controller.Publish)
	
	apiRouter.GET("/feed/", jwt.AuthWithoutLogin(), controller.Feed)

	apiRouter.GET("/publish/list/", jwt.Auth(), controller.PublishList)



	// extra apis - I
	apiRouter.POST("/favorite/action/", jwt.Auth(), controller.FavoriteAction)
	apiRouter.GET("/favorite/list/", jwt.Auth(), controller.GetFavouriteList)
	apiRouter.POST("/comment/action/", jwt.Auth(), controller.CommentAction)
	apiRouter.GET("/comment/list/", jwt.AuthWithoutLogin(), controller.CommentList)
	// extra apis - II
	apiRouter.POST("/relation/action/", jwt.Auth(), controller.RelationAction)
	apiRouter.GET("/relation/follow/list/", jwt.Auth(), controller.GetFollowing)
	apiRouter.GET("/relation/follower/list", jwt.Auth(), controller.GetFollowers)
}
