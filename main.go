package main

import (
	"github.com/streadway/amqp"
	"log"
	"rmq.jtthink.com/AppInit"
)

func main() {
	conn := AppInit.GetConn()
	defer conn.Close()
	c, err:=conn.Channel()  // 创建channel
	if err != nil{
		log.Fatal(err)
	}
	defer c.Close()

	// 创建队列
	queue,err := c.QueueDeclare("test",false,false,false,false,nil) // 测试队列
	if err != nil{
		log.Fatal(err)
	}

	// channel完成数据的传输，所以要调用channel
	err = c.Publish("",queue.Name,false,false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body: []byte("test002"),
		},
	) // 发布一条消息
	if err != nil{
		log.Fatal(err)
	}
	log.Println("发送消息成功")

}


