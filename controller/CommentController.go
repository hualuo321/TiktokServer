package controller

import (
	"TikTok/dao"
	"TikTok/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type CommentListResponse struct {
	StatusCode   int32                 `json:"status_code"`
	StatusMsg    string                `json:"status_msg,omitempty"`
	Comment_list []service.CommentInfo `json:"comment_list,omitempty"`
}

//-发表 or 删除评论 comment/action/
func Comment_Action(c *gin.Context) {
	//获取userId
	userId, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	//错误处理
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: -1,
			StatusMsg:  "comment userId json invalid",
		})
		return
	}
	//获取videoId
	videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	//错误处理
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: -1,
			StatusMsg:  "comment videoId json invalid",
		})
		return
	}
	//获取操作类型
	actionType, err := strconv.ParseInt(c.Query("action_type"), 10, 32)
	//错误处理
	if err != nil || actionType < 1 || actionType > 2 {
		c.JSON(http.StatusOK, Response{
			StatusCode: -1,
			StatusMsg:  "comment actionType json invalid",
		})
		return
	}
	//调用service层评论函数
	commentService := new(service.CommentServiceImpl)
	if actionType == 1 { //actionType为1，则进行发表评论操作
		content := c.Query("comment_text")
		var sendComment dao.Comment
		sendComment.User_id = userId
		sendComment.Video_id = videoId
		sendComment.Comment_text = content
		//发表评论
		if commentService.Send(sendComment) != nil { //发表评论失败
			c.JSON(http.StatusOK, Response{
				StatusCode: -1,
				StatusMsg:  "send comment failed",
			})
			return
		}

		/*//测试count函数+++++++++++++++++++++：
		count0, err := commentService.CountFromVideoId(videoId)
		if err != nil {
			c.JSON(http.StatusOK, Response{
				StatusCode: -1,
				StatusMsg:  "count comment failed",
			})
			return
		}
		var str string
		str = "send comment " + strconv.Itoa(int(count0)) + " success"
		//+++++++++++++++++++++++++++++++++++++*/

		//发表评论成功
		c.JSON(http.StatusOK, Response{
			StatusCode: 0,
			StatusMsg:  "send comment success",
		})
		return
	} else { //actionType为2，则进行删除评论操作
		//获取要删除的评论的id
		commentId, err := strconv.ParseInt(c.Query("comment_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, Response{
				StatusCode: -1,
				StatusMsg:  "delete commentId invalid",
			})
			return
		}
		//删除评论操作
		err = commentService.DelComment(commentId)
		if err != nil { //删除评论失败
			str := err.Error()
			c.JSON(http.StatusOK, Response{
				StatusCode: -1,
				StatusMsg:  str,
			})
			return
		}
		//删除评论成功
		c.JSON(http.StatusOK, Response{
			StatusCode: 0,
			StatusMsg:  "delete comment success",
		})
		return
	}
}

//-查看评论列表 comment/list/
func Comment_List(c *gin.Context) {
	//获取userId
	userId, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	//错误处理
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: -1,
			StatusMsg:  "comment userId json invalid",
		})
		return
	}
	//获取videoId
	videoId, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	//错误处理
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: -1,
			StatusMsg:  "comment videoId json invalid",
		})
		return
	}
	log.Println(videoId)
	log.Println(userId)
	//调用service层评论函数
	commentService := new(service.CommentServiceImpl)
	commentList, err := commentService.GetList(videoId, userId)
	if err != nil { //获取评论列表失败
		log.Println(err.Error())
		c.JSON(http.StatusOK, CommentListResponse{
			StatusCode: -1,
			StatusMsg:  "get comment list failed",
		})
		return
	}

	//测试获取了几条数据：
	//count := len(commentList)
	//var str string
	//str = "get comment list" + strconv.Itoa(count) + " success"
	//log.Println(str)

	//获取评论列表成功
	c.JSON(http.StatusOK, CommentListResponse{
		StatusCode:   0,
		StatusMsg:    "get comment list success",
		Comment_list: commentList,
	})
	//log.Println(str)
	return
}