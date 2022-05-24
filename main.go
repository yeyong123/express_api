/*
File Name: main.go
Created Date: 2022-05-24 10:59:04
Author: yeyong
Last modified: 2022-05-24 11:01:31
*/
package main
import (
    "senkoo.cn/express"
    "log"
)
func main() {
    /*
    测试圆通
    */
    res, err := express.QueryRouter("yto", "", "766666")
    if err != nil {
        log.Fatal(err)
    }
    log.Println(res)
}
