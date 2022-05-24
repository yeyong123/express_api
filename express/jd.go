/*
File Name: jd.go
Created Date: 2022-05-12 16:15:13
Author: yeyong
Last modified: 2022-05-17 14:43:57
*/
package express

import (
    "senkoo.cn/config"
    "time"
    "sort"
    "fmt"
    "net/url"
    "errors"
)

var jdConf = config.JdSetting

type Jd struct {
    appKey      string
    appSecret   string
    apiURL      string
    code        string
    callback    string
    token       string
}

func NewJD(customer_code string) *Jd {
    if len(customer_code) > 0 {
        customer_code = customer_code
    } else {
        customer_code = jdConf.CustomerCode
    }
    jd := &Jd{
        appKey: jdConf.JdAppKey,
        appSecret: jdConf.JdAppSecret,
        apiURL: jdConf.JdApiURL,
        code: customer_code,
        callback: "https://api.yunlu-crm.com/callback",
    }
    jd.GetToken("")
    return jd
}
func (jd *Jd) ScopeGetJDCode() string {
    tn := time.Now().UnixNano() / 1e6
    u := "https://open-oauth.jd.com/oauth2/to_login"
    apiURL := fmt.Sprintf("%s?app_key=%s&response_type=code&redirect_uri=%s&state=%d&scope=snsapi_base", u, jd.appKey, jd.callback, tn)
    fmt.Println("#", apiURL)
    return apiURL
}
func (jd *Jd) ReferToken() (string, error) {
    refer_token, err := CacheFetch("jd_refer_token")
    if err != nil {
        return "", err
    }
    u := "https://open-oauth.jd.com/oauth2/refresh_token"
    body := url.Values{
        "app_key": {jd.appKey},
        "app_secret": {jd.appSecret},
        "grant_type": {"refresh_token"},
        "refresh_token": {refer_token},
    }
    req := &ReqParam{
        URL: u,
        Body: body,
    }
    res, err := Request(req)
    if err != nil {
        return "", err
    }
    token := res["access_token"].(string)
    refer_token = res["refresh_token"].(string)
    CacheSet(map[string]interface{}{"jd_token": token, "jd_refer_token": refer_token}, time.Second * 3600000)
    jd.token = token
    return token, nil
}

func (jd *Jd) GetToken(code string) (string, error) {
    key := "jd_token"
    token, err := CacheFetch(key)
    if err == nil {
        ok, err := CacheTTL(key)
        if err != nil {
            return "", err
        }
        if ok {
            return jd.ReferToken()
        }
        jd.token = token
        return token, nil
    }
    if len(code) == 0 {
        return "", errors.New("无效的code")
    }
    u := "https://open-oauth.jd.com/oauth2/access_token"
    body := url.Values{
        "app_key": {jd.appKey},
        "app_secret": {jd.appSecret},
        "grant_type": {"authorization_code"},
        "code": {code},
    }
    param := &ReqParam {
        URL: u,
        Body: body,
    }
    res, err := Request(param)
    if err != nil {
        return "", err
    }
    fmt.Println("RESULT: ", res)
    if res["code"].(float64) != 0 {
        return "", errors.New("获取 token 失败")
    }
    token = res["access_token"].(string)
    refer_token := res["refresh_token"].(string)
    CacheSet(map[string]interface{}{key: token, "jd_refer_token": refer_token}, time.Second * 3600000)
    jd.token = token
    return token, nil
}

func (jd *Jd) SendOrder(params map[string]interface{}) (result map[string]interface{}, err error) {
    param := map[string]string {
        "salePlat": "0030001",
        "customerCode": jd.code,
        "orderId": "SKNTest200720114207",
        "senderName": "张三",
        "senderAddress": "北京亦庄经济开发区科创十一街与经海四路交叉口西北150米京东总部4号楼",
        "receiveName": "李四",
        "receiveAddress": "北京亦庄经济技术开发区科创十一街18号院京东总部1号楼",
        "senderMobile": "18621554897",
        "receiveMobile": "18621554897",
        "packageCount": "1",
        "weight": "2.5",
        "vloumn": "100",
    }
    method := "jingdong.ldop.waybill.receive"
    jd.requests(method, param)
    return 
}

func (jd *Jd) QueryRouter(no string) (result map[string]interface{}, err error) {
    if len(no) == 0 {
        return nil, errors.New("运单号为空")
    }
    params := map[string]string{
        "customerCode": jd.code,
        "waybillCode": no,
    }
    method := "jingdong.ldop.receive.trace.get"
    res, err := jd.requests(method, params)
    if err != nil {
        return nil, err
    }
    fmt.Println("-------", res)
    tmp := res["jingdong_ldop_receive_trace_get_responce"].(map[string]interface{})
    if tmp["code"].(string) != "0" {
        return nil, errors.New("未获取到快递信息")
    }
    tmp = tmp["querytrace_result"].(map[string]interface{})
    items := tmp["data"].([]interface{})
    if len(items) == 0 {
        return nil, errors.New("暂无物流追踪信息")
    }
    routers := []map[string]interface{}{}
    for _, t := range items {
        item := t.(map[string]interface{})
        ti, _ := time.Parse("2006/01/02 15:04:05", item["opeTime"].(string))
        temp := map[string]interface{}{
            "waybill_no": item["waybillCode"],
            "code": item["opeTitle"],
            "route": item["opeRemark"],
            "date": &ti,
        }
        routers = append(routers, temp)
    }
    result = map[string]interface{}{
        "waybill_no": no,
        "waybill_routers": routers,
    }
    return result, nil
}

func (jd *Jd) CancelOrder(no string) (result map[string]interface{}, err error) {
    method := "jingdong.ldop.delivery.provider.cancelWayBill"
    params := map[string]string {
        "waybillCode": no,
        "customerCode": jd.code,
        "source": "SKCRM",
        "cancelReason": "客户取消",
        "operatorName": "系统管理员",
    }
    res, err := jd.requests(method, params)
    if err != nil {
        return nil, err
    }
    tmp := res["jingdong_ldop_delivery_provider_cancelWayBill_responce"].(map[string]interface{})
    if tmp["code"].(string) != "0" {
        return nil, errors.New("取消无效")
    }
    tmp = tmp["responseDTO"].(map[string]interface{})
    if tmp["statusCode"].(float64) != 0 {
        return nil, errors.New(fmt.Sprintf("%v", tmp["statusMessage"]))
    }
    return tmp,  nil
}
func (jd *Jd) QueryPrice(data map[string]interface{}) (result map[string]interface{}, err error) {
    method := "jingdong.eclp.b2c.queryEstimatedFreights"
    params := map[string]string {
        "senderAddress": "上海市松江区千帆路288弄2 号楼 901",
        "receiverAddress": "北京市大兴区前门路22 号",
        "weight": "1",
        "businessType": "0",
        "orderTime": "2022-05-15 18:00:00",
    }
    res, err := jd.requests(method, params)
    if err != nil {
        return nil, err
    }
    tmp := res["jingdong_eclp_b2c_queryEstimatedFreights_responce"].(map[string]interface{})
    if tmp["code"].(string) != "0" {
        return nil, errors.New("无数据")
    }
    tmp = tmp["responseDTO"].(map[string]interface{})
    if tmp["statusCode"].(float64) != 0 {
        return nil, errors.New(fmt.Sprintf("v%", tmp["statusMessage"]))
    }
    items := tmp["data"].([]interface{})
    item := items[0].(map[string]interface{})
    result = map[string]interface{} {
        "total": item["freight"],
    }
    fmt.Println("-----------", result)
    return result, nil
}

func (jd *Jd) NotifyPickup(data map[string]interface{}) (result map[string]interface{}, err error) {
    return 
}

func (jd *Jd) requests(method string, params map[string]string) (result map[string]interface{}, err error) {
    t := ti()
    params["access_token"] = jd.token
    params["app_key"] = jd.appKey
    params["timestamp"] = t
    params["v"] = "2.0"
    params["method"] = method
    body := sortMap(params)
    body = fmt.Sprintf("%s%s%s", jd.appSecret, body, jd.appSecret)
    sign := SignMD5Upper(body)
    param := url.Values{}
    for k, v := range params {
        param.Set(k, v)
    }
    param.Set("sign", sign)
    reqBody := &ReqParam{
        URL: jd.apiURL,
        Body: param,
    }
    res, err := Request(reqBody)
    if err != nil {
        return nil, err
    }
    if res["error_response"] != nil {
        err_resp := res["error_response"].(map[string]interface{})
        if err_resp["code"].(string) != "0" {
            return nil, errors.New(fmt.Sprintf("%v", err_resp["zh_desc"]))
        }
    }
    return res, nil
}

func sortMap(data map[string]string) (res string) {
    keys := make([]string, 0, len(data))
    for k := range data {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    for _, k := range keys {
        res += fmt.Sprintf("%s%s", k, data[k])
    }
    return res
}
func ti() string {
    t := time.Now()
    y := t.Year()
    m := t.Month()
    d := t.Day()
    h := t.Hour()
    min := t.Minute()
    sec := t.Second()
    return fmt.Sprintf("%2d-%02d-%02d %02d:%02d:%02d", y, m, d, h, min, sec)
}
