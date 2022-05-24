/*
File Name: best.go
Created Date: 2022-05-13 17:28:42
Author: yeyong
Last modified: 2022-05-15 20:19:03
*/
package express

import (
    "senkoo.cn/config"
    "fmt"
    "errors"
    "net/url"
    "time"
)

var bConf = config.BestSetting

type Best struct {
    cid     string
    secret  string
    apiURL  string
}

func NewBest() *Best {
    return &Best{
        cid: bConf.BCustomerCode,
        secret: bConf.BSecret,
        apiURL: bConf.BApiURL,
    }
}


func (b *Best) SendOrder(params map[string]interface{}) (result map[string]interface{}, err error) {
    method := "KD_CREATE_ORDER_NOTIFY"
    sender := map[string]interface{} {
        "county": "松江区",
        "city": "上海",
        "prov": "上海",
        "address": "上海市松江区千帆路 288 弄 2 号楼 901",
        "mobile": "13876543451",
        "name": "张三",
    }
    receiver := map[string]interface{} {
        "county": "大兴区",
        "city": "北京",
        "prov": "北京",
        "address": "北京市大兴区千帆路 288 弄 2 号楼 901",
        "mobile": "13876543451",
        "name": "李四",
    }
    items := map[string]interface{} {
        "item": []map[string]interface{}{
            {"number": "1", "itemName": "海报"},
       },
    }
    params = map[string]interface{}{
            "txLogisticID": "SKN123123131",
            "orderType": "1",
            "serviceType": "1",
            "orderFlag": "1",
            "sendStartTime": "2022-05-15 18:00:00",
            "sendEndTime": "2022-05-15 20:00:00",
            "itemWeight": 5.0,
            "sender": sender,
            "receiver": receiver,
            "items": items,
    }
    b.requests(method, params)
    return 
}

func (b *Best) QueryRouter(no string) (result map[string]interface{}, err error) {
    method := "KD_TRACE_QUERY"
    params := map[string]interface{}{
        "mailNos": map[string]interface{}{
            "mailNo": []string{no},
        },
    }
    res, err := b.requests(method, params)
    if err != nil {
        return nil, err
    }
    tmp := res["traceLogs"].([]interface{})
    if len(tmp) == 0 {
        return nil, errors.New("无记录")
    }
    res = tmp[0].(map[string]interface{})
    items := res["traces"].(map[string]interface{})
    if len(items) == 0 {
        return nil, errors.New("无数据")
    }
    routers := []map[string]interface{}{}
    for _, t := range items["trace"].([]interface{}) {
        item := t.(map[string]interface{})
        var ti time.Time
        if item["acceptTime"] != nil {
            ti, _ = time.Parse("2006-01-02 15:04:05", item["acceptTime"].(string))
        }
        tmp := map[string]interface{}{
            "code": item["scanType"],
            "route": item["remark"],
            "waybill_no": no,
            "date": &ti,
        }
        routers  = append(routers, tmp)
    }
    result = map[string]interface{}{
        "waybill_no": no,
        "waybill_routers": routers,
    }
    fmt.Println("->", result)
    return
}

func (b *Best) UpdateOrder(no string) (result map[string]interface{}, err error) {
    method := "KD_UPDATE_ORDER_NOTIFY"
    params := map[string]interface{}{
        "txLogisticID": "SKN123123131",
        "itemWeight": 12.00,
    }
    b.requests(method, params)
    return
}


func (b *Best) DigestWaybill() (result map[string]interface{}, err error) {
    method := "KD_WAYBILL_APPLY_NOTIFY"
    list := []map[string]interface{} {
        {
            "sendMan": "张三",
            "sendManPhone": "18666667777",
            "sendManAddress": "上海市松江区千帆路 288 弄 2 号楼 901",
            "sendProvince": "上海",
            "sendCity": "上海",
            "sendCounty": "松江区",
            "receiveMan": "李四",
            "receiveManPhone": "13788898987",
            "receiveManAddress": "北京市大兴区阡陌路 19 号",
            "receiveProvince": "北京",
            "receiveCity": "北京",
            "receiveCounty": "大兴区",
            "txLogisticID": "SKN88990001231",
            "itemName": "广告物料",
            "itemWeight": 4.2,
            "itemCount": 3,
        },
    }
    auth := map[string]interface{}{
        "username": "DZMD",
        "pass": "800best",
    }
    params := map[string]interface{}{
        "deliveryConfirm": false,
        "EDIPrintDetailList": list,
        "msgId": "SKN1231231",
        "auth": auth,
    }
    res, err := b.requests(method, params)
    if err != nil {
        return nil, err
    }
    temp := res["EDIPrintDetailList"].([]interface{})
    result = temp[0].(map[string]interface{})
    return result, nil
}

func (b *Best) CancelOrder(no string) (result map[string]interface{}, err error) {
    method := "KD_CANCEL_ORDER_NOTIFY"
    params := map[string]interface{}{
        "txLogisticID": "SKN123123131",
        "reason": "疫情原因",
    }
    b.requests(method, params)
    return
}
func (b *Best) QueryPrice(data map[string]interface{}) (result map[string]interface{}, err error) {
    method := "KY_PRICE_AGING_SEARCH"
    params := map[string]interface{} {
        "scrCityCode": "330100000000",
        "destCityCode": "440100000000",
        "weight": "1.68",
    }
    res, err := b.requests(method, params)
    if err != nil {
        return nil, err
    }
    fmt.Println(res)
    return
}

func (b *Best) NotifyPickup(data map[string]interface{}) (result map[string]interface{}, err error) {
    return 
}

func (b *Best) StopExpress() error {
    method := "KD_WAYBILL_ADDRESS_ANOMALY"
    params := map[string]interface{} {
        "provinceName": "上海市",
        "cityName": "上海市",
        "countyName": "松江区",
        "address": "上海市松江区千帆路288弄",
    }
    res, err := b.requests(method, params)
    if err != nil {
        return err
    }
    fmt.Println("结果: ", res)
    return nil
}

func (b *Best) requests(method string, data map[string]interface{}) (map[string]interface{}, error) {
    sign, tmp := SignMd5(data, b.secret)
    params := url.Values{
        "bizData": {string(tmp)},
        "serviceType": {method},
        "partnerID": {b.cid},
        "sign": {sign},
    }
    req := &ReqParam{
        URL: b.apiURL,
        Body: params,
    }
    fmt.Println("---", req, params)
    res, err := Request(req)
    if err != nil {
        fmt.Println("ERR: ", err)
        return nil, err
    }
    if len(res) == 0 {
        return nil, errors.New("无数据")
    } 
    if !res["result"].(bool) {
        return nil, errors.New(fmt.Sprintf("查询错误: %v", res["remark"]))
    }
    return res, nil
}
