/*
File Name: zto.go
Created Date: 2022-05-12 16:15:01
Author: yeyong
Last modified: 2022-05-14 10:21:03
*/
package express
import (
    "senkoo.cn/config"
    "fmt"
)

var ztoConf = config.ZtoSetting
    
type ZTO struct {
    appSecret   string
    appKey      string
    apiURL      string
}

func NewZTO() *ZTO {
    return &ZTO{
        appSecret: ztoConf.ZtoAppSecret,
        appKey: ztoConf.ZtoAppKey,
        apiURL: ztoConf.ZtoApiURL,
    }
}


func (zto *ZTO) SendOrder(params map[string]interface{}) (result map[string]interface{}, err error) {
    params = map[string]interface{} {
        "partnerType": "2",
        "orderType": "1",
        "partnerOrderCode": "SKN0012312312",
        "accountInfo": map[string]interface{}{
            "accountId": "test",
            "customerId": "GPG1576724269",
            "type": "1",
        },
        "senderInfo": map[string]interface{}{
            "senderName": " 张三",
            "senderAddress": "上海市松江区千帆路288弄2号楼901",
            "senderDistrict": "松江区",
            "senderProvince": "上海市",
            "senderCity": "上海市",
            "senderMobile": "18976545451",
        },
        "receiveInfo": map[string]interface{}{
            "receiverAddress": "大兴区刘家庄路 18 弄",
            "receiverCity": "北京市",
            "receiverName": "李四",
            "receiverDistrict": "大兴区",
            "receiverProvince": "北京市",
            "receiverMobile": "13778766789",
        },
        "summaryInfo": map[string]interface{}{
            "quantity": 2,
            "startTime": "2022-05-13 12:16:12",
            "endTime": "2022-05-13 20:00:00",
        },
    }
    zto.requests("zto.open.createOrder", params)
    return
}

func (zto *ZTO) QueryRouter(no string) (result map[string]interface{}, err error) {
    params := map[string]interface{}{
        "billCode": no,
    }
    zto.requests("zto.open.getRouteInfo", params)
    return
}

func (zto *ZTO) CancelOrder(no string) (result map[string]interface{}, err error) {
    params := map[string]interface{}{
        "billCode": "73100012775188",
        "cancelType": "2",
        "orderCode": "220513000005234105",
    }
    zto.requests("zto.open.cancelPreOrder", params)
    return
}
func (zto *ZTO) QueryPrice(data map[string]interface{}) (result map[string]interface{}, err error) {
    return
}

func (zto *ZTO) NotifyPickup(data map[string]interface{}) (result map[string]interface{}, err error) {
    return 
}

func (z *ZTO) requests(method string, params map[string]interface{}) error {
    urlc := fmt.Sprintf("%s%s", z.apiURL, method)
    sign, data := SignMd5ToBase64(params, z.appSecret)
    param := &ReqParam {
        URL: urlc,
        Data: data,
        Type: "json",
        Header: map[string]string{
            "x-appKey": z.appKey,
            "x-dataDigest": sign,
        },
    }
    res, err := Request(param)
    if err != nil {
        return err
    }
    fmt.Println(res)
    return nil
}
