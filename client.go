package main

import (
	"fmt"
	"log"
	"rmq.jtthink.com/AppInit"
)

func main()  {
	conn := AppInit.GetConn()
	defer conn.Close()
	c, err:=conn.Channel()  // 创建channel
	if err != nil{
		log.Fatal(err)
	}
	defer c.Close()
	// 队列名称、
	msgs,err := c.Consume("test", "c1", false, false, false, false, nil)
	if err != nil{
		log.Fatal(err)
	}

	for msg := range msgs{
		//msg.Ack()
		fmt.Println(msg.DeliveryTag, string(msg.Body)) // 每条消息的唯一标识
	}
}


