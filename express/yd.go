/*
File Name: yd.go
Created Date: 2022-05-12 16:15:29
Author: yeyong
Last modified: 2022-05-15 19:10:24
*/

//韵达快递接口

package express

import (
    "senkoo.cn/config"
    "fmt"
    "time"
    "errors"
)

var ydConf = config.YdSetting

type YD struct {
    appKey      string
    appSecret   string
    apiURL      string
    partnerId   string
    secret      string
}

func NewYD() *YD {
    return &YD {
        appKey: ydConf.YdAppKey,
        appSecret: ydConf.YdAppSecret,
        apiURL: ydConf.YdApi,
        partnerId: "201700101001", //需要从合作网点获取
        secret: "123456789", //需要从合作网点获取
    }
}

func (yd *YD) SendOrder(params map[string]interface{}) (result map[string]interface{}, err error) {
    //下单接口, 电子面单下单接口
    orders := []map[string]interface{}{
        {
            "order_serial_no": "SKN0098788",
            "khddh": "SKN0098788",
            "weight": 1.0,
            "order_type": "common",
            "node_id": "350",
            "items": []map[string]interface{} {
                {
                    "name": "广告物料",
                    "number": 5,
                    "remark": "广告物料",
                },
            },
            "sender": map[string]interface{}{
                "name": "张三",
                "address": "上海市松江区千帆路 288 弄 2 号楼 901",
                "phone": "18987766542",
            },
            "receiver": map[string]interface{} {
                "name": "李四",
                "address": "北京市大兴区前门路 98弄 12 号",
                "phone": "13798766667",
            },
        },
    }
    item := map[string]interface{} {
        "appid": yd.appKey,
        "partner_id": yd.partnerId,
        "secret": yd.secret,
        "orders": orders,
    }
    method := "openapi-api/v1/accountOrder/createBmOrder"
    res, err := yd.requests(method, item)
    if err != nil {
        return nil, err
    }
    result = map[string]interface{}{}
    tmp := res["data"].([]interface{})
    for _, tt := range tmp {
        item := tt.(map[string]interface{})
        oid := item["order_serial_no"].(string)
        result[oid] = item["mail_no"].(string)
        result["data"] = item["pdf_info"]
    }
    return result, nil
}

func (yd *YD) QueryRouter(no string) (result map[string]interface{}, err error) {
    result = make(map[string]interface{})
    params := map[string]interface{} {
        "mailno": no,
    }
    method := "openapi/outer/logictis/query"
    res, err := yd.requests(method, params)
    if err != nil {
        return result, err
    }
    tmp := res["data"].(map[string]interface{})
    if tmp["result"].(string) == "false" {
        return nil, errors.New("无数据")
    }
    waybill_no := tmp["mailno"].(string)
    items := tmp["steps"].([]interface{})
    routers := []map[string]interface{}{}
    for _, tmp := range items {
        item := tmp.(map[string]interface{})
        temp := map[string]interface{} {
            "waybill_no": waybill_no,
            "code": actionKey[item["action"].(string)],
            "route": item["description"].(string),
            "date": ParseTime(item["time"].(string)),
        }
        routers = append(routers, temp)
    }
    result = map[string]interface{} {
        "mailNo": waybill_no,
        "waybill_routers": routers,
    }
    return result, nil
}
func ParseTime(timeStr string) *time.Time {
    ti, _ := time.Parse("2006-01-02 15:04:05", timeStr)
    return &ti
}

func (yd *YD) CancelOrder(no string) (result map[string]interface{}, err error) {
    orders := []map[string]interface{} {
        {
            "order_serial_no": "SKN0098788",
            "mailno": "312006166679665",
        },
    }
    item := map[string]interface{} {
        "appid": yd.appKey,
        "partner_id": yd.partnerId,
        "secret": yd.secret,
        "orders": orders,
    }
    method := "openapi-api/v1/accountOrder/cancelBmOrder"
    _, err = yd.requests(method, item)
    if err != nil {
        return nil, err
    }
    result = map[string]interface{}{
        "oid": no,
    }
    return result, nil
}
func (yd *YD) QueryPrice(data map[string]interface{}) (result map[string]interface{}, err error) {
    data = map[string]interface{} {
        "startCity": "上海",
        "endCity": "北京",
        "weight": "1",
    }
    method := "openapi-api/v1/order/getFreightInfo"
    res, err := yd.requests(method, data)
    result = map[string]interface{} {
        "total": res["data"],
    }
    fmt.Println("==", result)
    return result, nil
}

func (yd *YD) NotifyPickup(data map[string]interface{}) (result map[string]interface{}, err error) {
    return 
}

func (yd *YD) requests(method string, params map[string]interface{}) (result map[string]interface{}, err error) {
    urlc := fmt.Sprintf("%s%s", yd.apiURL, method)
    tnow := fmt.Sprintf("%d", time.Now().UnixNano() / 1e6)
    sec := fmt.Sprintf("_%s",yd.appSecret)
    sign, data := SignMd5(params, sec)
    req := &ReqParam {
        URL: urlc,
        Type: "json",
        Data: data,
        Header: map[string]string{
            "app-key": yd.appKey,
            "sign": sign,
            "req-time": tnow,
        },
    }
    res, err := Request(req)
    if err != nil {
        return nil, err
    }
    if !res["result"].(bool)  {
        return nil, errors.New(fmt.Sprintf("%v", res["message"]))
    }
    return res, nil
}

var actionKey = map[string]string {
    "ACCEPT":	"收件扫描",
    "GOT":	"揽件扫描",
    "ARRIVAL": "入中转",
    "DEPARTURE":	"出中转",
    "SENT":	"派件中",
    "INBOUND":	"第三方代收入库",
    "SIGNED":	"已签收",
    "OUTBOUND": "第三方代收快递员取出",
    "SIGNFAIL": "签收失败",
    "RETURN":	"退回件",
    "ISSUE": "问题件",
    "REJECTION": "拒收",
    "OTHER": "其他",
    "OVERSEA_IN":	"入库扫描",
    "OVERSEA_OUT":	"出库扫描",
    "CLEARANCE_START":	"清关开始",
    "CLEARANCE_FINISH":	"清关结束",
    "CLEARANCE_FAIL":	"清关失败",
    "OVERSEA_ARRIVAL":	"干线到达",
    "OVERSEA_DEPARTURE": "干线离开",
    "TRANSFER": "转单",
}
