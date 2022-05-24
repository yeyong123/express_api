/*
File Name: sto.go
Created Date: 2022-05-13 17:27:33
Author: yeyong
Last modified: 2022-05-15 19:09:43
*/
package express
import (
    "senkoo.cn/config"
    "fmt"
    "time"
    "net/url"
    "errors"
    "strconv"
)

var stoConf = config.StoSetting

type STO struct {
    key         string
    secret      string
    code        string
    api         string
    number      string
}

func NewSto() *STO {
    return &STO {
        key: stoConf.StoAppKey,
        secret: stoConf.StoAppSecret,
        api: stoConf.StoApiURL,
        code: stoConf.StoCode,
        number: stoConf.StoNumber,
    }
}

func (s *STO) SendOrder(params map[string]interface{}) (result map[string]interface{}, err error) {
    method := "OMS_EXPRESS_ORDER_CREATE"
    tcode := "sto_oms"
    sender := map[string]interface{}{
        "name": "张三",
        "mobile": "13899987876",
        "province": "上海",
        "city": "上海",
        "area": "松江",
        "address": "上海松江区千帆路 288 弄 2 号楼 901",
    }
    receiver := map[string]interface{} {
         "name": "李四",
        "mobile": "13899187876",
        "province": "北京",
        "city": "北京",
        "area": "大兴区",
        "address": "北京市大兴区前门路288 弄 2 号楼 901",
    }
    cargo := map[string]interface{} {
        "battery": "30",
        "goodsType": "小件",
        "goodsName": "广告物料",
    }
    customer :=  map[string]interface{} {
        "siteCode": "666666",
        "customerName": "666666000001",
        "sitePwd": "abc123",
    }
    params = map[string]interface{}{
        "orderNo": "SKN09922231",
        "orderSource": s.number,
        "billType": "00",
        "orderType": "01",
        "sender": sender,
        "receiver": receiver,
        "cargo": cargo,
        "customer": customer,
    }
    res, err := s.requests(method, tcode, params)
    if err != nil {
        return nil, err
    }
    tmp := res["data"].(map[string]interface{})
    fmt.Println("--", tmp)
    return tmp, nil
}

func (s *STO) QueryRouter(no string) (result map[string]interface{}, err error) {
    method := "STO_TRACE_QUERY_COMMON"
    tcode := "sto_trace_query"
    params := map[string]interface{}{
        "waybillNoList": []string{no},
    }
    res, err := s.requests(method, tcode, params)
    if err != nil {
        return nil, err
    }
    tmp := res["data"].(map[string]interface{})
    items := tmp[no].([]interface{})
    routers := []map[string]interface{}{}
    for _, t := range items {
        item := t.(map[string]interface{})
        ti, _ := time.Parse("2006-01-02 15:04:05", item["opTime"].(string))
        et := map[string]interface{}{
            "code": item["scanType"],
            "route": item["memo"],
            "waybill_no": no,
            "date": &ti,
        }
        routers = append(routers, et)
    }
    result = map[string]interface{}{
        "mailNo": no,
        "waybill_routers": routers,
    }
    fmt.Println("-", result)
    return result, nil
}

func (s *STO) CancelOrder(no string) (result map[string]interface{}, err error) {
    method := "EDI_MODIFY_ORDER_CANCEL"
    tcode := "edi_modify_order"
    params := map[string]interface{} {
        "billCode": no,
        "orderSource": s.number,
        "orderType": "01",
    }
    res, err := s.requests(method, tcode, params)
    if err != nil {
        return nil, err
    }
    fmt.Println("--", res)
    return res, nil
}
func (s *STO) QueryPrice(data map[string]interface{}) (result map[string]interface{}, err error) {
    method := "QUERY_SEND_SERVICE_DETAIL"
    tcode := "ORDERMS_API"
    params := map[string]interface{}{
        "SendName": "张三",
        "SendMobile": "13899876546",
        "SendProv": "上海",
        "SendCity": "上海",
        "SendArea": "松江区",
        "SendAddress": "上海市松江区千帆路 288 弄 2 号楼 901",
        "RecName": "李四",
        "RecMobile": "13677765423",
        "RecProv": "北京",
        "RecCity": "北京",
        "RecArea": "大兴区",
        "RecAddress": "北京市大兴区前门路 2 号",
        "OpenId": "SKN987667",
        "Weight": "1",
    }
    res, err := s.requests(method, tcode, params)
    if err != nil {
        return nil, err
    }
    if res["success"].(string) != "true" {
        return nil, errors.New("为获取价格信息")
    }
    data = res["data"].(map[string]interface{})
    tmp := data["AvailableServiceItemList"].([]interface{})
    total := ""
    for _, t := range tmp {
        item := t.(map[string]interface{})
        v, ok := item["feeModel"].(map[string]interface{})
        if !ok {
            continue
        }
        total = v["totalPrice"].(string)
    }
    ntemp, _ := strconv.Atoi(total)
    total = fmt.Sprintf("%.2f", float64(ntemp  / 100))
    result = map[string]interface{}{
        "total": total,
    }
    return result, nil
}

func (s *STO) NotifyPickup(data map[string]interface{}) (result map[string]interface{}, err error) {
    return 
}

func (s *STO) requests(method, tcode string, params map[string]interface{}) (result map[string]interface{}, err error) {
    sign, tmp := SignMd5ToBase64(params, s.secret)
    param := url.Values{
        "content": {string(tmp)},
        "data_digest": {sign},
        "api_name": {method},
        "from_appkey": {s.key},
        "from_code": {s.code},
        "to_appkey": {tcode},
        "to_code": {tcode},
    }
    req := &ReqParam{
        URL: s.api,
        Body: param,
    }
    res, err := Request(req)
    return res, err
}
