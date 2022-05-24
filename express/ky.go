/*
File Name: ky.go
Created Date: 2022-02-14 16:31:10
Author: yeyong
Last modified: 2022-05-24 10:51:50
*/
package express

import (
    "bytes"
    "crypto/md5"
    "encoding/base64"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "senkoo.cn/config"
    "strings"
    "time"
)

var kyConf = config.KySetting

type KuaYue struct {
    appkey string
    appsecret string
    timestamp string
    token string
    tokenURL string
    apiURL string
    account string
    flag string
}
//sender, recevier : =[companyName, person, phone, provinceName, cityName, address]
type ConcatInfo struct {
    CompanyName     string  `json:"companyName"`
    Person          string  `json:"person"`
    Phone           string  `json:"phone"`
    ProvinceName    string  `json:"provinceName"`
    CityName        string  `json:"cityName"`
    Address         string  `json:"address"`
}

func NewKuayue(account string) *KuaYue {
    st := time.Now().UnixNano() / 1e6
    tstap := fmt.Sprintf("%d", st)
    apiURL := fmt.Sprintf("%s%s", kyConf.KyURL, "/router/rest")
    fmt.Println(kyConf)
    ky := &KuaYue{
        appkey: kyConf.KyAppKey,
        appsecret: kyConf.KyAppSecret,
        tokenURL: kyConf.KyTokenURL,
        apiURL: apiURL,
        timestamp: tstap,
        account: account,
        flag: kyConf.KyFlag,
    }
    err := ky.getToken()
    if err != nil {
        return nil
    }
    return ky
}

func (ky *KuaYue) getToken() error {
    key := "ky_token"
    val, err := CacheFetch(key)
    if err == nil {
        ky.token = val
        return nil
    }

    params := map[string]interface{}{
        "appkey": ky.appkey,
        "appsecret": ky.appsecret,
    }
    data, _ := json.Marshal(params)
    param := &ReqParam{
        URL: ky.tokenURL,
        Type: "json",
        Data: data,
    }
    tmp, err := Request(param)
    if err != nil {
        return err
    }
    if tmp["code"].(float64) != 0 {
        return errors.New("获取 token 错误")
    }
    resp := tmp["data"].(map[string]interface{})
    ky.token = resp["token"].(string)
    expire := int(resp["expire_time"].(float64))
    CacheSet(map[string]interface{}{key: resp["token"]}, time.Second *  time.Duration(expire))
    return nil
}

func (ky *KuaYue) signMD5(params map[string]interface{}) string {
    tmp, _ := json.Marshal(params)
    str := fmt.Sprintf("%s%s%s", ky.appsecret, ky.timestamp, string(tmp))
    temp := md5.New()
    temp.Write([]byte(str))
    str = hex.EncodeToString(temp.Sum(nil))
    return strings.ToUpper(str)
}



func (ky *KuaYue) SendOrder(data map[string]interface{}) (result map[string]interface{}, err error) {
    //sender, recevier : =[companyName, person, phone, provinceName, cityName, address]
    //goodsInfo := [][length, width, height, count, goodsCode, goodsName, goodsWeight, goodsVolume,
    var items = data["data"].([]map[string]interface{})
    var add = data["add"].(string)

    /*
        10-表示当天达，20-表示次日达，30-表示隔日达，40-表示陆运件，50-表示同城次日，
        60-表示次晨达，70-表示同城即日，80-表示航空件，160-表示省内次日，170-表示省内即日，
        210-表示空运，220-表示专运（传代码）
    */
    var tmps = []map[string]interface{}{}
    for _, item := range items {
        basicParams := map[string]interface{}{
            "serviceMode": 30,
            "payMode": 10,
            "paymentCustomer": ky.account,
            "goodsType": "广告物料",
            "receiptFlag": 30,
            "count": item["count"],
            "subscriptionService": "10",
            "actualWeight": item["weight"],
            "additionalService":"20",
        }
        if add == "20"{
            basicParams["additionalService"] = add
            basicParams["count"] = item["add_cnt"]
        }


        basicParams["orderId"] = item["orderId"]
        basicParams["preWaybillDelivery"] = item["sender"]
        basicParams["preWaybillPickup"] = item["receiver"]
        basicParams["preWaybillGoodsSizeList"] = item["goodsInfo"]
        tmps = append(tmps, basicParams)
    }
    params := map[string]interface{} {
        "repeatCheck": "20",
        "customerCode": ky.account,
        "platformFlag": ky.flag,
        "orderInfos": tmps,
        //"orderInfos": []map[string]interface{} {
        //    {
        //        "preWaybillDelivery": data["sender"],
        //        "preWaybillPickup": data["receiver"],
        //        "serviceMode": 20, //10当天达, 20次日达,30隔日达, 40陆运件,50同城次日,60次晨达,70 同城即日,80 航空件,160 省内次日, 170 省内即日,210 空运
        //        "payMode": 10,
        //        "paymentCustomer": ky.account,
        //        "goodsType": "广告物料",
        //        "count": 1,
        //        "orderId": data["orderId"],
        //        "receiptFlag": 20,
        //        //"pictureSubscription": 10, //图片订阅
        //        //additionalService: 10, //10=打印子母件, 40绑定子母件
        //        //subscriptionService: 10, //路由订阅服务
        //    },
        //},
    }

    tmp, _ := json.Marshal(params)
    sign := ky.signMD5(params)
    res, err := ky.requests(ky.apiURL, tmp, sign,"open.api.openCommon.planOrder")
    if err != nil {
        return nil, err
    }  
    if int(res["code"].(float64)) != 10000 {
        return nil, errors.New(fmt.Sprintf("%v", res["msg"]))
    } 

    temp := res["data"].([]interface{})
    if len(temp) == 0 {
        return nil, errors.New(fmt.Sprintf("%v", res["msg"]))
    }
    result = map[string]interface{}{}
    for _, item := range temp {
        tmpk := item.(map[string]interface{})
        oid := tmpk["orderId"].(string)
        wno := tmpk["waybillNumber"].(string)
        result[oid] = wno
    }
    result["data"] = res["data"]
    return result, nil
}

func (ky *KuaYue) CancelOrder(no string) (map[string]interface{}, error) {
    //取消订单
    params := map[string]interface{}{
        "customerCode": ky.account,
        "waybillNumber": no,
        "platformFlag": ky.flag,
    }
    tmp, _ := json.Marshal(params)
    sign := ky.signMD5(params)

    res, err := ky.requests(ky.apiURL, tmp, sign, "open.api.openCommon.cancelOrder")
    if err != nil {
        return nil, err
    }   
    return res, nil
}


func (ky *KuaYue) QueryWaybillPicture(no string) (string, *bytes.Buffer,error) {
    /*
    运单图片下载
    */

    params := map[string]interface{}{
        "customerCode":ky.account,
        "waybillNumber":no,
        "pictureType":"60",
    }
    tmp, _ := json.Marshal(params)
    sign := ky.signMD5(params)
    res, err := ky.requests(ky.apiURL, tmp, sign, "open.api.openCommon.queryWaybillPicture")
    if err != nil {
        return "",nil, err
    }
    if res["code"].(float64) == 10000 {
        val := res["data"].(map[string]interface{})
        temp := val["filePictureInfoRes"].([]interface{})
        dst := temp[0].(map[string]interface{})

        fname := dst["originalName"].(string)
        data := dst["picture"].(string)
        d, _ := base64.StdEncoding.DecodeString(data)
        b := bytes.NewBuffer(d)
        return fname,b,nil
    }
    return "",nil,err
}




func (ky *KuaYue) NotifyPickup(data map[string]interface{}) (map[string]interface{}, error) {
    //通知取货
    data["customerCode"] =  ky.account
    tmp, _ := json.Marshal(data)
    sign := ky.signMD5(data)
    res, err := ky.requests(ky.apiURL, tmp, sign, "open.api.openCommon.notifyPickupGoods")
    if err != nil {
        return nil, err
    }
    if res["code"].(float64) != 10000 {
        return nil, errors.New(fmt.Sprintf("%v", res["msg"]))
    }
    return res, nil
}

func (ky *KuaYue) QueryRouter(no string) (map[string]interface{}, error) {
    params := map[string]interface{} {
        "waybillNumbers": []string{no},
    }
    tmp, _ := json.Marshal(params)
    sign := ky.signMD5(params)
    res, err := ky.requests(ky.apiURL, tmp, sign, "open.api.openCommon.queryPublicRoute")
    if err != nil {
        return nil, err
    }
    if res["code"].(float64) != 10000 {
        return nil, err
    }
    data := res["data"]
    if data == nil {
        return nil, errors.New("无数据")
    }
    data1 := data.(map[string]interface{})
    res1 := data1["esWaybill"].([]interface{})
    if len(res1) == 0 {
        return nil, errors.New("无数据")
    }
    res2 := res1[0].(map[string]interface{})
    mailNo := res2["waybillNumber"]
    date := res2["expectedDate"]
    items := []map[string]interface{}{}
    for _, item := range res2["exteriorRouteList"].([]interface{}) {
        tmp1 := item.(map[string]interface{})
        dt := tmp1["uploadDate"].(string)
        ti, _ := time.Parse("2006-01-02 15:04:05", dt)
        temp := map[string]interface{}{
            "waybill_no": mailNo,
            "code": tmp1["routeStep"],
            "route": tmp1["routeDescription"],
            "date": &ti,
        }
        items = append(items, temp)
    }
    result := map[string]interface{} {
        "mailNo": mailNo,
        "expected_date": date,
        "waybill_routers": items,
    }
    return result, nil
}

func (ky *KuaYue) QueryPrice(data map[string]interface{}) (result map[string]interface{},  err error) {
    params := map[string]interface{}{
        "customerCode": ky.account,
        "serviceType": "40", //40: 陆运, 20//次日达
        "billingTime": "2022-06-12 14:00",
        "pickupCustomerCode": ky.account,
        "weight": data["weight"],
        "volume": data["volume"],
        "beginCityName": data["start_city"],
        "endCityName": data["end_city"],
        "beginAreaCode": "021",
        "endAreaCode": "010",
    }
    fmt.Println("ky account",ky.account,ky.appkey,ky.appsecret,ky.apiURL)
    tmp, _ := json.Marshal(params)
    sign := ky.signMD5(params)
    res, err := ky.requests(ky.apiURL, tmp, sign, "open.api.openCommon.queryFreightCharge")
    if err != nil {
        return nil, err
    }
    return res, nil
}

func (ky *KuaYue) QueryWaybillStatus(no string) (map[string]interface{}, error) {
    params := map[string]interface{} {
        "waybillNumber": no,
        "page": 1,
        "pageSize": 10,
        "customerCode": ky.account,
    }
    tmp, _ := json.Marshal(params)
    sign := ky.signMD5(params)
    res, err := ky.requests(ky.apiURL, tmp, sign, "open.api.openCommon.getWaybillBaseInfo")
    if err != nil {
        return nil, err
    }
    return res, nil
}

func (ky *KuaYue) BindPrinter() error {
    params := map[string]interface{} {
        "companyNo": ky.account,
        "printers": []map[string]interface{} {
            {
                "printerId": "Y1000018151904",
                "netType": "2",
            },
        },
        "marketer": "张三",
        "marketerNo": "000001",
    }
    tmp, _ := json.Marshal(params)
    sign := ky.signMD5(params)
    _, err := ky.requests(ky.apiURL, tmp, sign, "KYE_BindDevice")
    if err != nil {
        return err
    }
    return nil

}

func (ky *KuaYue) PrinterWaybill(no string) error {
    params := map[string]interface{} {
        "customerCode": ky.account,
        "platformFlag": ky.flag,
        "templateSizeType": "0",
        "waybillNumberInfos": []map[string]interface{} {
            {"waybillNumber": no},
        },
        "printClientId": "0AAD-9PK3-3KP9-DAA0",
    }
    tmp, _ := json.Marshal(params)
    sign := ky.signMD5(params)
    _, err := ky.requests(ky.apiURL, tmp, sign, "open.api.openCommon.print")
    if err != nil {
        return err
    }
    return nil

}

func (ky *KuaYue) PrinterList() (map[string]interface{}, error) {
    params := map[string]interface{} {
        "companyNo": ky.account,
    }

    tmp, _ := json.Marshal(params)
    sign := ky.signMD5(params)
    res, err := ky.requests(ky.apiURL, tmp, sign, "KYE_GetPrinters")
    if err != nil {
        return nil, err
    }
    return res, nil

}

func (ky *KuaYue) PrinterStatus() error {
    params := map[string]interface{} {
        "companyNo": ky.account,
        "printerNo": "Y1000018151904",
    }

    tmp, _ := json.Marshal(params)
    sign := ky.signMD5(params)
    res, err := ky.requests(ky.apiURL, tmp, sign, "KYE_GetPrinters")
    fmt.Println("####################", res, err)
    if err != nil {
        return err
    }
    return nil

}

func (ky *KuaYue) QueryTimelines(t, origin, dest string) (result []map[string]interface{}, err error) {
    /*
    查询预计能到的时效
    t: 发货的时间
    origin: 始发地
    desc: 目的地
    */
    method := "open.api.openCommon.queryTimeliness"
    params := map[string]interface{}{
        "customerCode": ky.account,
        "mailingTime": t,
        "sendAddress": origin,
        "collectAddress": dest,
    }
    tmp, _ := json.Marshal(params)
    sign := ky.signMD5(params)
    res, err := ky.requests(ky.apiURL, tmp, sign, method)
    if err != nil {
        return nil, err
    }
    temp, _ := json.Marshal(res["data"])
    json.Unmarshal(temp, &result)
    fmt.Println(string(temp))
    return result, nil
}


func (ky *KuaYue) requests(url string,  data []byte, sign, method string) (result map[string]interface{}, err error) {
    headers := map[string]string{
        "appkey": ky.appkey,
        "format": "json",
        "timestamp": ky.timestamp,
        "method": method,
        "sign": sign,
        "token": ky.token,
        "x-from": "openapi_app",
    }
    param := &ReqParam{
        Header: headers,
        URL: url,
        Data: data,
        Type: "json",
    }
    result, err = Request(param)
    if err != nil {
        return nil, err
    }
    if result["success"] == nil {
        return nil, errors.New(fmt.Sprintf("%v", result["msg"]))
    }
    if !result["success"].(bool) {
        return nil, errors.New(fmt.Sprintf("%v", result["msg"]))
    }
    return result, nil
}





