package rabbitmq

import (
	"TikTok/dao"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"strconv"
)

type CommentMQ struct {
	RabbitMQ
	channel   *amqp.Channel	// 消息队列通道
	queueName string		// 队列名称
	exchange  string		// 交换机名称
	key       string		// 路由键
}

// 获取CommentMQ的对应队列
func NewCommentRabbitMQ(queueName string) *CommentMQ {
	commentMQ := &CommentMQ{
		RabbitMQ:  *Rmq,
		queueName: queueName,	// 设置队列名称
	}

	cha, err := commentMQ.conn.Channel()	// 创建 RabbitMQ 通道
	commentMQ.channel = cha
	Rmq.failOnErr(err, "获取通道失败")
	return commentMQ
}

// Publish Comment的发布配置。
func (c *CommentMQ) Publish(message string) {
	// 声明消息要发送到的队列
	_, err := c.channel.QueueDeclare(
		c.queueName,	// 队列名称
		false,			// 是否持久化
		false,			// 是否为自动删除
		false,			// 是否具有排他性
		false,			// 是否阻塞
		nil,			// 额外属性
	)
	if err != nil {
		panic(err)
	}
	// 将消息发布到声明的队列
	err1 := c.channel.Publish(
		c.exchange,		// 交换机名称
		c.queueName,	// 队列名称
		false,			// 是否强制
		false,			// 是否立即
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err1 != nil {
		panic(err)
	}
}

// Consumer follow关系的消费逻辑。
func (c *CommentMQ) Consumer() {

	_, err := c.channel.QueueDeclare(c.queueName, false, false, false, false, nil)

	if err != nil {
		panic(err)
	}

	//2、接收消息
	msg, err := c.channel.Consume(
		c.queueName,
		//用来区分多个消费者
		"",
		//是否自动应答
		true,
		//是否具有排他性
		false,
		//如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
		false,
		//消息队列是否阻塞
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	//只有删除逻辑
	forever := make(chan bool)
	go c.consumerCommentDel(msg)

	//log.Printf("[*] Waiting for messages,To exit press CTRL+C")

	<-forever

}

// 数据库中评论删除的消费方式。
func (c *CommentMQ) consumerCommentDel(msg <-chan amqp.Delivery) {
	for d := range msg {
		// 参数解析，只有一个评论id
		cId := fmt.Sprintf("%s", d.Body)
		commentId, _ := strconv.Atoi(cId)
		//log.Println("commentId:", commentId)
		//删除数据库中评论信息
		err := dao.DeleteComment(int64(commentId))
		if err != nil {
			log.Println(err)
		}
	}
}

var RmqCommentDel *CommentMQ

// InitCommentRabbitMQ 初始化rabbitMQ连接。
func InitCommentRabbitMQ() {
	RmqCommentDel = NewCommentRabbitMQ("comment_del")
	go RmqCommentDel.Consumer()
}
