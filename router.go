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
	// 5.1 Gin 路由组监听获取已发布视频事件
	apiRouter.GET("/publish/list/", jwt.Auth(), controller.PublishList)
	// 6.1 Gin 路由组监听拉取视频列表事件
	apiRouter.GET("/feed/", jwt.AuthWithoutLogin(), controller.Feed) 
	// extra apis - I
	// 通过路由组来鉴权并调用FavoriteAction函数，实现点赞功能
	apiRouter.POST("/favorite/action/", jwt.Auth(), controller.FavoriteAction)
	// 通过路由组来鉴权并调用GetFavouriteList函数，实现获取点赞列表功能
	apiRouter.GET("/favorite/list/", jwt.Auth(), controller.GetFavouriteList)
	// 通过路由组来鉴权并调用CommentAction函数，实现评论功能
	apiRouter.POST("/comment/action/", jwt.Auth(), controller.CommentAction)
	// 通过路由组来鉴权并调用CommentList函数，实现拉取评论列表功能
	apiRouter.GET("/comment/list/", jwt.AuthWithoutLogin(), controller.CommentList)
	// extra apis - II
	// 通过路由组来鉴权并调用RelationAction函数，实现关注功能
	apiRouter.POST("/relation/action/", jwt.Auth(), controller.RelationAction)
	// 通过路由组来鉴权并调用GetFollowing函数，获取当前用户的关注列表
	apiRouter.GET("/relation/follow/list/", jwt.Auth(), controller.GetFollowing)
	// 通过路由组来鉴权并调用GetFollowers函数，获取当前用户的粉丝列表
	apiRouter.GET("/relation/follower/list", jwt.Auth(), controller.GetFollowers)
}
