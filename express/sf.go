/*
File Name: sf.go
Created Date: 2022-02-11 13:28:17
Author: yeyong
Last modified: 2022-05-24 10:50:12
*/
package express
import (
    "encoding/json"
    "net/url"
    "fmt"
    "time"
    "errors"
    "senkoo.cn/config"
)
/*
操作码(opCode)	操作名称(opName)
50	顺丰已收件
51	一票多件的子件
54	上门收件
30	快件在【XXX营业点】已装车,准备发往 【XXX集散中心】
31	快件到达 【XXX集散中心】
302	车辆发车
304	离开经停点
3036	快件在XXX ,准备送往下一站
33	派件异常原因
34	滞留件出仓
36	封车操作
10	办事处发车/中转发车/海关发车/机场发货
105	航空起飞
106	航空落地
11	办事处到车/中转到车/海关到车/机场提货
122	加时区域派件出仓
123	快件正送往顺丰店/站
125	快递员派件至丰巢
126	快递员取消派件将快件取出丰巢
127	寄件客户将快件放至丰巢
128	客户从丰巢取件成功
129	快递员从丰巢收件成功
130	快件到达顺丰店/站
131	快递员从丰巢收件失败
135	信息记录
136	落地配装车
137	落地配卸车
138	用户自助寄件
14	货件已放行
140	国际件特殊通知
141	预售件准备发运
147	整车在途
15	海关查验
151	分配师傅
152	师傅预约
153	师傅提货
154	师傅上门
16	正式报关待申报
17	海关待查
18	海关扣件
186	KC装车
187	KC卸车
188	KC揽收
189	KC快件交接
190	无人车发车
201	准备拣货
202	出库
205	仓库内操作-订单审核
206	仓库内操作-拣货
208	代理交接
211	星管家派件交接
212	星管家派送
214	星管家收件
215	星管家退件给客户
405	船舶离港
406	船舶到港
407	接驳点收件出仓
41	交收件联(二程接驳收件)
43	收件入仓
46	二程接驳收件
47	二程接驳派件
570	铁路发车
571	铁路到车
604	CFS清关
605	运力抵达口岸
606	清关完成
607	代理收件
611	理货异常
612	暂存口岸待申报
613	海关放行待补税
614	清关时效延长
619	检疫查验
620	检疫待查
621	检疫扣件
623	海关数据放行
626	到转第三方快递
627	寄转第三方快递
630	落地配派件出仓
631	新单退回
632	储物柜交接
633	港澳台二程接驳收件
634	港澳台二程接驳派件
64	晨配转出
642	门市/顺丰站快件上架
643	门市/顺丰站快件转移
646	包装完成
647	寄方准备快件中
648	快件已退回/转寄,新单号为: XXX
649	代理转运
65	晨配转入
651	SLC已揽件
655	合作点收件
656	合作点交接给顺丰
657	合作点从顺丰交接
658	合作点已派件
66	中转批量滞留
660	合作点退件给客户
664	客户接触点交接
676	顺PHONE车
677	顺手派
678	快件到达驿站
679	驿站完成派件
70	由于XXX原因 派件不成功
700	拍照收件
701	一票多件拍照收件
72	标记异常
75	混建包复核
77	中转滞留
830	出库
831	入库
833	滞留件入仓
84	重量复核
843	集收入库
844	配送出库
847	二程接驳
850	集收
851	集收
86	晨配装车
87	晨配卸车
870	非正常派件
88	外资航空起飞
880	上门派件
89	外资航空落地
900	订舱路由
921	晨配在途
930	外配装车
931	外配卸车
932	外配交接
933	外配出仓
934	外配异常
935	外配签收
950	快速收件
96	整车发货
97	整车签收
98	代理路由信息
99	应客户要求,快件正在转寄中
44	正在派送途中,请您准备签收(派件人:XXX,电话:XXX)
204	派件责任交接
80	已签收,感谢使用顺丰,期待再次为您服务
8000	在官网"运单资料&签收图",可查看签收人信
*/

var sfConf = config.SfSetting

type Sf struct {
    URL         string
    Checkword   string
    Account     string
    ClientCode string
    token       string
    timestamp   string
    reqID       string
}

 
func NewSf(account string) *Sf {
    reqId := "sk12322"
    tstap := time.Now().Unix()
    stamp := fmt.Sprintf("%d", tstap)
    sf := &Sf{
        URL: sfConf.SfURL,
        Checkword: sfConf.SfCheckword,
        Account: account,
        ClientCode: sfConf.SfClientCode,
        timestamp: stamp,
        reqID: reqId,
    }
    sf, err := sf.GetToken()
    if err != nil {
        fmt.Println("ERROR", err)
        return nil
    }
    return sf
}


func (sf *Sf) SendOrder(data map[string]interface{}) (result map[string]interface{}, err error) {
    //
    /*
    msgData := map[string]interface{} {
       "language": "zh-CN",
       "orderId": data["order_id"],
       "monthlyCard": sf.Account,
       "cargoDetails": []map[string]interface{}{
           {"count": 4, "unit": "个", "weight": 4, "name": "广告物料"},
       },
       "contactInfoList": []map[string]interface{}{
           {"contactType": 1, "company": "释空", "concat": "小白", "tel":"1387776565", "country": "CN", "province": "上海", "city": "上海市", "address": "上海市松江区千帆路 28 弄"},
           {"contactType": 2, "company": "释空", "concat": "小白", "tel":"1387776565", "country": "CN", "province": "上海", "city": "上海市", "address": "上海市松江区千帆路 28 弄"},
       },
    }

    */
    result = map[string]interface{}{}
    msgData := map[string]interface{} {
        "language": "zh_CN",
        "orderId": data["order_id"],
        "monthlyCard": sf.Account,
    }
    msgData["cargoDetails"] = data["cargo"]
    msgData["contactInfoList"] = data["customer"]
    datap, _ := json.Marshal(msgData)
    tmp, err := sf.requests(string(datap), "EXP_RECE_CREATE_ORDER")
    if err != nil {
        return nil, err
    }
    t1 := tmp["msgData"].(map[string]interface{})
    t2 := t1["waybillNoInfoList"].([]interface{})
    if len(t2) == 0 {
        return nil, errors.New("未获取到单号")
    }
    t3 := t2[0].(map[string]interface{})
    result["waybill_no"] = t3["waybillNo"].(string)
    result["order_id"] = t1["orderId"].(string)
    result["data"] = tmp
    return result, nil
}

func (sf *Sf) QueryRouter(no string) (data map[string]interface{}, err error) {
    code := "EXP_RECE_SEARCH_ROUTES"
    msgData := map[string]interface{} {
        "language": "zh-CN",
        "trackingType": 1,
        "trackingNumber": no,
    }
    datap, _ := json.Marshal(msgData)
    tmp, err := sf.requests(string(datap), code)
    if err != nil {
        return nil, err
    }
    resp := tmp["msgData"].(map[string]interface{})["routeResps"]
    resp1 := resp.([]interface{})
    if len(resp1) == 0 {
        return nil, errors.New("未查到数据")
    }
    resp2 := resp1[0].(map[string]interface{})
    mailNo := resp2["mailNo"].(string)
    items := resp2["routes"].([]interface{})
    kemp := []map[string]interface{}{}
    for _, tmp := range items {
        item := tmp.(map[string]interface{})
        dt := item["acceptTime"].(string)
        ti, _ := time.Parse("2006-01-02 15:04:05", dt)
        k1 := map[string]interface{}{
            "code": item["opCode"],
            "waybill_no": mailNo,
            "route": item["remark"],
            "date": &ti,
        }
        kemp = append(kemp, k1)
    }
    result2 := map[string]interface{}{
        "waybill_no": mailNo,
        "expected_date": "0",
        "waybill_routers": kemp,
    }
    return result2, nil
}


func (sf *Sf) GetToken() (*Sf, error) {
    key := "sf_token"
    val, err := CacheFetch(key)
    if err == nil {
        sf.token = val
        return sf, nil
    }
    urlp := "https://sfapi.sf-express.com/oauth2/accessToken"
    body := url.Values{
        "partnerID": {sf.ClientCode},
        "secret": {sf.Checkword},
        "grantType": {"password"},
    }
    req := &ReqParam{
        URL: urlp,
        Body: body,
    }
    tmp, err := Request(req)
    if err != nil {
        return nil, err
    }
    if tmp["apiResultCode"] != "A1000" {
        return nil, errors.New("获取 token 失败")
    }
    sf.token = tmp["accessToken"].(string)
    CacheSet(map[string]interface{}{key: tmp["accessToken"]}, time.Second * 7200)
    return sf, nil
}

func (sf *Sf) CancelOrder(no string) (map[string]interface{}, error) {
    return nil, nil
}

func (sf *Sf) QueryPrice(data map[string]interface{}) (map[string]interface{},  error) {
    code := "EXP_RECE_QUERY_DELIVERTM"
    msgData := map[string]interface{} {
        "businessType": "2",
        "weight": data["weight"],
        "volume": data["volume"],
        "searchPrice": "1",
        "destAddress": data["dest"],
        "srcAddress": data["src"],
    }

    datap, _ := json.Marshal(msgData)
    tmp, err := sf.requests(string(datap), code)
    if err != nil {
        return nil, err
    }
    if tmp["errorCode"].(string) != "S0000" {
        return nil, errors.New(fmt.Sprintf("%v", tmp["errorMsg"]))
    }
    tmp1 := tmp["msgData"].(map[string]interface{})
    tmp2 := tmp1["deliverTmDto"].([]interface{})
    tmp1 = tmp2[0].(map[string]interface{})
    return tmp1, nil
}


func (sf *Sf) NotifyPickup(data map[string]interface{}) (map[string]interface{}, error) {
    return nil, nil
}

func (sf *Sf) requests(msgData, code string) (result map[string]interface{}, err error) {
    params := url.Values{
        "partnerID": {sf.ClientCode},
        "requestID": {sf.reqID},
        "timestamp": {sf.timestamp},
        "serviceCode": {code},
        "accessToken": {sf.token},
        "msgData": {msgData},
    }
    req := &ReqParam{
        URL: sf.URL,
        Body: params,
    }
    tmp, err := Request(req)
    if err != nil {
        return nil, err
    }
    if tmp["apiResultCode"].(string) != "A1000" {
        return nil, errors.New(fmt.Sprintf("%v", tmp["apiErrorMsg"]))
    }
    resp := tmp["apiResultData"].(string)
    json.Unmarshal([]byte(resp), &result)
    if !result["success"].(bool) {
        return nil, errors.New(fmt.Sprintf("%v", result["errorMsg"]))
    }
    if result["errorCode"].(string) != "S0000" {
        return  nil, errors.New(fmt.Sprintf("%v", result["errorMsg"]))
    }
    return result, nil
}

