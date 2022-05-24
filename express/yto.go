/*
File Name: yto.go
Created Date: 2022-05-12 16:14:41
Author: yeyong
Last modified: 2022-05-14 10:23:17
*/
package express

import (
    "senkoo.cn/config"
    "time"
    "fmt"
    "encoding/json"
)
var yconf = config.YtoSetting

type YTO struct {
    appKey      string
    appSecret   string
    apiURL      string
}

func NewYto() *YTO {
    return &YTO {
        appKey: yconf.YtoCode,
        appSecret: yconf.YtoSecret,
        apiURL: yconf.YtoApi,
    }
}

func (yto *YTO) SendOrder(params map[string]interface{}) (result map[string]interface{}, err error) {
    urlc := fmt.Sprintf("%s%s","open/korder_create_adapter/v1/C7ZyZv/", yto.appKey)
    data := map[string]interface{} {
        "logisticsNo": "SKN9980999",
        "senderName": "张三",
        "senderProvinceName": "上海",
        "senderCityName": "上海",
        "senderAddress": "上海市松江区千帆路 288 弄 2 号楼 901",
        "senderMobile": "13588876545",
        "recipientName": "李四",
        "recipientProvinceName": "北京",
        "recipientCityName": "北京",
        "recipientAddress": "北京市大兴区前门路 28 号",
        "recipientMobile": "18987676545",
        "remark": "广告物料-海报",
        "startTime": "2022-05-13 15:12:12",
        "endTime": "2022-05-13 20:00:00",
        "weight": 4.0,
        "cstOrderNo": "SKN9980999",
    }
    method := "korder_create_adapter"
    yto.requests(urlc, method, data)
    return 
}

func (yto *YTO) QueryRouter(no string) (result map[string]interface{}, err error) {
    params := map[string]interface{}{
        "Number": no,
    }
    method := "track_query_adapter"
    urlc := fmt.Sprintf("%s%s", "open/track_query_adapter/v1/C7ZyZv/",yto.appKey)
    yto.requests(urlc, method, params)
    return
}

func (yto *YTO) CancelOrder(no string) (result map[string]interface{}, err error) {
    params := map[string]interface{} {
        "logisticsNo": "SKN9980999",
        "cancelDesc": "疫情原因",
    }
    method := "korder_cancel_adapter"
    urlc := fmt.Sprintf("%s%s", "open/korder_cancel_adapter/v1/C7ZyZv/",yto.appKey)
    yto.requests(urlc, method, params)
    return
}
func (yto *YTO) QueryPrice(data map[string]interface{}) (result map[string]interface{}, err error) {
    return
}

func (yto *YTO) NotifyPickup(data map[string]interface{}) (result map[string]interface{}, err error) {
    return 
}

func (yto *YTO) requests(urlc, method string, params map[string]interface{}) (result map[string]interface{}, err error) {
    t := time.Now().UnixNano() / 1e6
    sec := fmt.Sprintf("%s%s%s", method, "v1", yto.appSecret)
    sign, tmp := SignMd5ToBase64(params, sec)
    data := map[string]interface{}{
        "timestamp": fmt.Sprintf("%d",t),
        "param": string(tmp),
        "format": "json",
        "sign": sign,
    }
    temp, _ := json.Marshal(data)
    urlc = fmt.Sprintf("%s%s", yto.apiURL, urlc)
    param := &ReqParam{
        URL: urlc,
        Type: "json",
        Data: temp,
    }
    res, err := Request(param)
    if err != nil {
        return nil, err
    }
    fmt.Println("#=>", res)
    return res, nil
}
