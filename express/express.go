/*
File Name: express.go
Created Date: 2022-02-11 13:24:18
Author: yeyong
Last modified: 2022-05-24 11:01:18
*/
package express

import (
    "crypto/md5"
    "encoding/hex"
    "encoding/json"
    "net/url"
    "bytes"
    "io/ioutil"
    "net/http"
    "crypto/tls"
    "encoding/base64"
    "errors"
    "fmt"
    "time"
    "senkoo.cn/service"
    "strings"
)

var redis = service.Redis

type Express interface {
    SendOrder(data map[string]interface{}) (map[string]interface{}, error)
    QueryRouter(no string) (map[string]interface{}, error)
    CancelOrder(no string) (map[string]interface{}, error)
    QueryPrice(data map[string]interface{}) (map[string]interface{}, error)
    NotifyPickup(data map[string]interface{}) (map[string]interface{}, error)
}

var ExpressName = map[string]string{
    "jd": "京东快递",
    "yd": "韵达快递",
    "zto": "中通快递",
    "sto": "申通快递",
    "yto": "圆通快递",
    "sf": "顺丰速运",
    "ky": "跨越速运",
    "best": "百世快递",
    "deppon": "德邦快递",
}

func NewWaybill(kind, account string) Express {
    //account 月结账号
    var express Express
    switch kind {
    case "sf":
        if len(account) == 0 {
            account = sfConf.SfAccount
        }
        sf := NewSf(account)
        express = sf
    case "ky":
        if len(account) == 0 {
            account = kyConf.KyAccount
        }
        ky := NewKuayue(account)
        express = ky
    case "jd":
        if len(account) == 0 {
            account = jdConf.CustomerCode
        }
        jd := NewJD(account)
        express = jd
    case "best":
        express = NewBest()
    case "sto":
        express = NewSto()
    case "yto":
        express = NewYto()
    case "zto":
        express = NewZTO()
    case "yd":
        express = NewYD()
    }
    return express
}

func SendOrder(kind, account string, data map[string]interface{}) (map[string]interface{}, error) {
    ex := NewWaybill(kind, account)
    if ex == nil {
        return nil, errors.New("无效请求")
    }
    return ex.SendOrder(data)
}

func QueryRouter(kind, account, no string) (map[string]interface{}, error) {
    ex := NewWaybill(kind, account)
    res, err := ex.QueryRouter(no)
    if err != nil {
        return nil, err
    }
    return res, err
}

func QueryPrice(kind, account string, data map[string]interface{}) (map[string]interface{}, error) {
    ex := NewWaybill(kind, account)
    return ex.QueryPrice(data)
}


func CancelOrder(kind, account, no string) (map[string]interface{}, error) {
    ex := NewWaybill(kind, account)
    return ex.CancelOrder(no)
}

func NotifyPickup(kind, account string, data map[string]interface{}) (map[string]interface{}, error) {
    ex := NewWaybill(kind, account)
    return ex.NotifyPickup(data)
}

func sign_md5(val string) []byte {
    m := md5.New()
    m.Write([]byte(val))
    return m.Sum(nil)
}


func SignMD5Upper(tmp string) string {
    m := sign_md5(tmp)
    str := hex.EncodeToString(m)
    return strings.ToUpper(str)
}

func SignMd5(data map[string]interface{}, secret string) (string, []byte){
    temp, _ := json.Marshal(data)
    itemStr := fmt.Sprintf("%s%s", string(temp), secret)
    m := sign_md5(itemStr)
    return hex.EncodeToString(m), temp
}

func SignHexAndBase64(params map[string]interface{}, secret string) (string, []byte) {
    str, tmp := SignMd5(params, secret)
    return base64.StdEncoding.EncodeToString([]byte(str)), tmp
}

func SignMd5ToBase64(data map[string]interface{}, secret string) (string, []byte) {
    temp, _ := json.Marshal(data)
    itemStr := fmt.Sprintf("%s%s", string(temp), secret)
    m := sign_md5(itemStr)
    return base64.StdEncoding.EncodeToString(m), temp
}

type ReqParam struct {
    URL         string
    Type        string
    Method      string
    Data        []byte
    Body        url.Values
    Header      map[string]string
}

func Request(param *ReqParam) (result map[string]interface{}, err error) {
    var req *http.Request
    method := "POST"
    if len(param.Method) > 0 {
        method = strings.ToUpper(param.Method)
    }
    ctype := "application/x-www-form-urlencoded;charset=UTF-8"
    if param.Type == "json" {
        ctype = "application/json"
        req, err = http.NewRequest(method, param.URL, bytes.NewBuffer(param.Data))
        if err != nil {
            fmt.Println("HTTP ERROR: ", err)
            return nil, err
        }
    } else {
        req, err = http.NewRequest(method, param.URL, strings.NewReader(param.Body.Encode()))
        if err != nil {
            fmt.Println("HTTP ERROR: ", err)
            return nil, err
        }
    }
    req.Header.Set("Content-Type", ctype)
    for k, v := range param.Header {
        req.Header.Add(k, v)
    }
    client := &http.Client {
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        },
    }
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return nil, errors.New("HTTP 请求失败")
    }
    resBody, _ := ioutil.ReadAll(resp.Body)
    err = json.Unmarshal(resBody, &result)
    if err != nil {
        return nil, err
    }
    return result, nil
}

func CacheSet(data map[string]interface{}, ex time.Duration) error {
    for k, v := range data {
        item, _ := json.Marshal(v)
        redis.Set(service.Ctx, k, item, ex)
    }
    return nil
}

func CacheFetch(token string) (result string, err error) {
    if len(token) == 0 {
        return "", errors.New("key不能为空")
    }
    val, ok := redis.Get(service.Ctx, token).Result()
    if ok != nil {
        return "", errors.New("查询为空")
    }
    json.Unmarshal([]byte(val), &result)
    return result, nil
}

func CacheTTL(key string) (bool, error) {
    //检查当前的Key 还有多久过期
    val, err := redis.TTL(service.Ctx, key).Result()
    if err != nil {
        return false, err
    }
    if val < 60 * time.Second {
        return true, nil
    }
    return false, nil
}

