package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
	"log"
	"rmq.jtthink.com/Lib"
	"rmq.jtthink.com/Trans"
)
var MyCron *cron.Cron

func initCron() error {
	MyCron = cron.New(cron.WithSeconds())  // 支持秒级定时器
	_, err := MyCron.AddFunc("0/3 * * * * *", FailTrans)
	if err != nil{
		return err
	}
	_, err2 := MyCron.AddFunc("0/4 * * * * *", BackMoney)
	if err2 != nil{
		return err2
	}
	_,err3:=MyCron.AddFunc("0/2 * * * * *", reSendMsg)//重发
	if err3!=nil{
		return err3
	}
	return nil
}

const FailSql = "update translog set status = 2 where status=0 and TIMESTAMPDIFF(SECOND,updatetime,now())>20"
const BackSql = "select tid, `from`, money from translog where status = 2 and isback=0 limit 10"
const resendSql="select * from translog where status=0 and TIMESTAMPDIFF(SECOND,updatetime,now())<=8 "

// 定时取消交易
func FailTrans()  {
	_,err := Trans.GetDB().Exec(FailSql)
	if err != nil{
		log.Println(err)
	}
}

// 取出交易时间在8秒内, 且status=0的数据 进行mq重发
func reSendMsg(){
	rows,err:=Trans.GetDB().Queryx(resendSql)
	if err!=nil{
		log.Println(err)
	}
	transList:=[]*Trans.TransModel{}
	err=sqlx.StructScan(rows,&transList)
	if err!=nil{
		log.Println(err)
	}else{
		mq:=Lib.NewMQ()
		defer mq.Channel.Close()
		for _,tm:=range transList{
			jsonb,_:=json.Marshal(tm) //可以使用协程
			err=mq.SendMessage(Lib.ROUTER_KEY_TRANS,Lib.EXCHANGE_TRANS,string(jsonb))
			if err!=nil{
				log.Println(err)
			}else {
				log.Println("重发成功:",tm)
			}
		}
	}
}

// 定时任务协程化 很可能第一个协程没完成 第二个协程又开始了
// 所以会造成数据的脏读, 单协程 --- 使用锁

var islock = false
func clearTx(tx *sqlx.Tx)  {  // 清理事务
	err := tx.Commit()
	// 事务已经结束了 commit会出错,所以做如下判断 这种错误就不是错误
	if  err != nil && err != sql.ErrTxDone{
		// 真的出错
		log.Println("tx error", err)
	}
	islock = false
}

// 还钱
func BackMoney()  {
	if islock{
		log.Println("locked.return.....")
		return
	}
	tx, err := Trans.GetDB().BeginTxx(context.Background(), nil)
	if err != nil{
		log.Println("事务失败",err)
		return
	}
	islock = true // 加锁
	//time.Sleep(time.Second * 6)
	defer clearTx(tx) // 清理事务
	// 只需要关注rollback,出错就回滚即可
	rows,err := tx.Queryx(BackSql)
	if err != nil{
		tx.Rollback()
		return
	}
	defer rows.Close()
	tms := []Trans.TransModel{}
	_ = sqlx.StructScan(rows, &tms)
	for _, tm := range tms{
		_,err=tx.Exec("update usermoney set user_money=user_money+? where user_name=?",tm.Money,tm.From)
		if err!=nil{
			tx.Rollback()
		}
		_,err=tx.Exec("update translog set isback=1 where tid=?",tm.Tid)
		if err!=nil{
			tx.Rollback()
		}
	}
}


func main()  {
	c := make(chan error)
	go func() {
		err:=Trans.DBInit("a")
		if err!=nil{
			c<-err
		}
	}()

	// 是独立部署的，虽然代码写在这边在一起了 需要单独部署 和httpserver是分开的
	// 分2个容器运行
	go func() {
		err := initCron()
		if err != nil{
			c <- err
		}
		MyCron.Start() // 开启定时任务
	}()

	err := <-c
	log.Fatal(err)
}
