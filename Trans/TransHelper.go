package Trans

import (
	"context"
	"fmt"
	"log"
)

// 转账 针对A公司
func TransMoney(tm *TransModel) error {
	// must begin tx一旦出错,就会panic,被中间件拦截到
	// context定义超时时间,超时就抛出异常
	tx:=GetDB().MustBeginTx(context.Background(), nil)
	ret,_:= tx.NamedExec(`update usermoney set user_money=user_money-:money 
    	  where user_name=:from and user_money>=:money
	`, tm)
	rowAffected,_ := ret.RowsAffected()
	if rowAffected == 0 {
		// sql语句失败了
		err := tx.Rollback()
		if err != nil{
			log.Println(err)
		}
		return fmt.Errorf("扣款失败")
	}
	ret,_ = tx.NamedExec("insert into translog(`from`,`to`,money,updatetime) values(:from,:to,:money,now())", tm)
	rowAffected,_ = ret.RowsAffected()
	if rowAffected == 0 {
		// sql语句失败了
		err := tx.Rollback()
		if err != nil{
			log.Println(err)
		}
		return fmt.Errorf("插入日志失败")
	}
	tid, _:=ret.LastInsertId() // 赋值tid, tid交易号  a产生的
	tm.Tid = tid
	return tx.Commit()
}

