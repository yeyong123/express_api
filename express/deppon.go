/*
File Name: deppon.go
Created Date: 2022-05-15 20:27:05
Author: yeyong
Last modified: 2022-05-17 11:45:30
*/
package express

import (
    "senkoo.cn/config"
    "net/url"
    "fmt"
    "time"
    "errors"
)

var dConf = config.DepponSetting

type Deppon struct {
    key             string
    secret          string
    api             string
    code            string
    stamp           string
    customerCode    string
    now             string
}

func NewDeppon(code string) *Deppon {
    tnow := time.Now()
    now := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", tnow.Year(), tnow.Month(), tnow.Day(), tnow.Hour(), tnow.Minute(), tnow.Second())
    customer_code := dConf.DepponCustomerCode
    if len(code) > 0 {
        customer_code = code
    }
    return &Deppon {
        key: dConf.DepponKey,
        secret: dConf.DepponSecret,
        api: dConf.DepponApi,
        code: dConf.DepponCode,
        customerCode: customer_code,
        stamp: fmt.Sprintf("%d", (int(time.Now().UnixNano()  / 1e6))),
        now: now,
    }
}

func (d *Deppon) SendOrder(data map[string]interface{}) (result map[string]interface{}, err error) {
    method := "dop-standard-ewborder/createOrderNotify.action"
    sender := map[string]interface{} {
        "name": "张三",
        "mobile": "13899987656",
        "province": "上海",
        "city": "上海",
        "county": "松江区",
        "address": "上海市松江区千帆路 288 弄 2 号楼 901",
    }

    receiver := map[string]interface{} {
        "name": "李四",
        "mobile": "13899987656",
        "province": "北京",
        "city": "北京",
        "county": "大兴区",
        "address": "北京市大兴区阡陌路2号901",
    }
    packageInfo := map[string]interface{} {
        "cargoName": "广告物料",
        "totalNumber": 1,
        "totalWeight": 1,
        "deliveryType": 3,
    }
    service := map[string]interface{} {
        "addServices": "2",
    }
    code := "SKN102927638884"
    params := map[string]interface{}{
        "logisticID": fmt.Sprintf("%s%s", d.secret, code),
        "payType": "2",
        "gmtCommit": d.now,
        "sendStartTime": "2022-05-16 18:00:00",
        "sendEndTime": "2022-05-16 20:00:00",
        "custOrderNo": code,
        "needTraceInfo": "0", //是否需要推送物流状态
        "companyCode": d.code,
        "orderType": "3",
        "transportType": "PACKAGE",
        "customerCode": d.customerCode, //子母件219401
        "sender": sender,
        "receiver": receiver,
        "packageInfo": packageInfo,
        "addServices": service,
    }
    res, err := d.requests(method, params)
    if err != nil {
        return nil, err
    }
    if res["result"].(string) != "true" {
        return nil, errors.New(fmt.Sprintf("%v", res["reason"]))
    }
    waybill_no := res["mailNo"]
    result = map[string]interface{} {
        "order_id": code,
        "waybill_no": waybill_no,
    }
    fmt.Println("--", result)
    return result, nil
}
func (d *Deppon) CancelOrder(no string) (result map[string]interface{}, err error) {
    method := "standard-order/cancelOrder.action"
    params := map[string]interface{} {
        "mailNo": no,
        "cancelTime": d.now,
        "remark": "客户取消",
    }
    res, err := d.requests(method, params)
    if err != nil {
        return nil, err
    }
    if res["result"].(string) != "true" {
        return nil, errors.New(fmt.Sprintf("%v", res["reason"]))
    }
    result = map[string]interface{} {
        "waybill_no": no,
    }
    return result, nil
}
func (d *Deppon) QueryRouter(no string) (result map[string]interface{}, err error) {
    method := "standard-order/newTraceQuery.action"
    params := map[string]interface{} {
        "mailNo": "479000244251",
    }
    res, err := d.requests(method, params)
    if err != nil {
        return nil, err
    }
    if res["result"].(string) != "true" {
        return nil, errors.New(fmt.Sprintf("%v", res["reason"]))
    }
    tmp := res["responseParam"].(map[string]interface{})
    items := tmp["trace_list"].([]interface{})
    fmt.Println("#", items)
    wno := tmp["tracking_number"]
    routers := []map[string]interface{}{}
    for _, t := range items {
        item := t.(map[string]interface{})
        ti, _ := time.Parse("2006-01-02 15:04:05", item["time"].(string))
        temp := map[string]interface{} {
            "waybill_no": no,
            "code": depponStatus[item["status"].(string)],
            "route": item["description"],
            "date": &ti,
        }
        routers = append(routers, temp)
    }
    result = map[string]interface{} {
        "mailNo": wno,
        "waybill_routers": routers,
    }
    return result, nil
}
func (d *Deppon) QueryPrice(data map[string]interface{}) (result map[string]interface{}, err error) {
    method := "standard-order/queryPriceTime.action"
    params := map[string]interface{}{
        "originalsStreet": "北京-北京市-大兴区,前门路",
        "originalsaddress": "上海-上海市-松江区,千帆路299弄",
        "sendDateTime": "2022-05-18 12:12:00",
        "totalVolume": "1",
        "totalWeight": "1.23",
    }
    res, err := d.requests(method, params)
    if err != nil {
        return nil, err
    }
    fmt.Println(res)
    return
}

func (d *Deppon) TrackingRoute(no string) error {
    method := "standard-order/standTraceSubscribe.action"
    params := map[string]interface{} {
        "tracking_number": no,
        //"order_number": "SKN102927638884",
    }
    res, err := d.requests(method, params)
    if err != nil {
        return err
    }
    fmt.Println("###################", res)
    return nil
}

func (d *Deppon) requests(method string, params map[string]interface{}) (result map[string]interface{}, err error) {
    params["logisticCompanyID"] = "DEPPON"
    sec := fmt.Sprintf("%s%s", d.key, d.stamp)
    sign, tmp := SignHexAndBase64(params, sec)
    param := url.Values{
        "params": {string(tmp)},
        "digest": {sign},
        "timestamp": {d.stamp},
        "companyCode": {d.code},
    }
    api := fmt.Sprintf("%s%s", d.api, method)
    req := &ReqParam{
        URL: api,
        Body: param,
    }
    return Request(req)
}

var depponStatus = map[string]interface{} {
    "GOT": "开单",
    "DEPARTURE": "出站",
    "ARRIVAL": "进站",
    "ADVANCE_DELIVERY": "预派送",
    "SENT_SCAN":"派送",
    "ERROR": "滞留,延时派送",
    "FAILED": "客户拒签/运单作废",
    "SIGNED": "签收",
    "BACK_SIGNED": "退回件签收",
    "OPERATETRACK": "转寄",
    "STA_INBOUND":"快递员入柜",
    "STA_SIGN":"用户提货（快递柜签收）",
}
