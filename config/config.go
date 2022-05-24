/*
File Name: config.go
Created Date: 2022-05-24 10:53:42
Author: yeyong
Last modified: 2022-05-24 11:01:04
*/
package config

import (
    "gopkg.in/ini.v1"
    "log"
    "os"
    "path/filepath"
)

type RedisConf struct {
    RedisAddr, RedisPass string
    RedisDB int
}


type KyConf struct {
    KyAccount, KyAppKey, KyFlag, KyTokenURL, KyURL, KyAppSecret string
}

type SfConf struct {
    SfCheckword, SfClientCode, SfURL, SfAccount string
}

type WeatherConf struct {
    WeatherAPI, WeatherSecret string
}

type AppConf struct {
    ServicePort string
}

type ZtoConf struct {
    ZtoAppSecret, ZtoAppKey, ZtoApiURL string
}

type YtoConf struct {
    YtoCode, YtoSecret, YtoApi string
}

type YdConf struct {
    YdApi       string
    YdAppKey    string
    YdAppSecret string
}

type JDConf struct {
    JdAppKey        string
    JdAppSecret     string
    JdApiURL        string
    CustomerCode    string
}

type StoConf struct {
    StoApiURL       string
    StoAppKey       string
    StoAppSecret    string
    StoCode         string
    StoNumber       string
}

type BestConf struct {
    BCustomerCode   string
    BSecret         string
    BApiURL         string
}

type DepponConf struct {
    DepponCode      string
    DepponSecret    string
    DepponApi       string
    DepponKey       string
    DepponCustomerCode string
}

var CacheSetting = &RedisConf{}
var KySetting = &KyConf{}
var SfSetting = &SfConf{}
var WeatherSetting = &WeatherConf{}
var ZtoSetting = &ZtoConf{}
var YtoSetting = &YtoConf{}
var YdSetting = &YdConf{}
var JdSetting   = &JDConf{}
var BestSetting = &BestConf{}
var StoSetting = &StoConf{}
var DepponSetting = &DepponConf{}

func init() {
    env := os.Getenv("GO_ENV")
    section := "config_dev.ini"
    if len(env) > 0 && env == "prod" {
        section = "config_dev.ini"
    }
    dirPath := filepath.Join("./", section)
    cfg, err := ini.Load(dirPath)
    if err != nil {
        log.Fatal("未获取到配置文件")
    }
    err = cfg.Section("cache").MapTo(CacheSetting)
    if err != nil {
        log.Fatal("初始化缓存 Redis 失败")
    }
    cfg.Section("ky").MapTo(KySetting)
    cfg.Section("sf").MapTo(SfSetting)
    cfg.Section("yto").MapTo(YtoSetting)
    cfg.Section("zto").MapTo(ZtoSetting)
    cfg.Section("yundax").MapTo(YdSetting)
    cfg.Section("weather").MapTo(WeatherSetting)
    cfg.Section("jd").MapTo(JdSetting)
    cfg.Section("best").MapTo(BestSetting)
    cfg.Section("sto").MapTo(StoSetting)
    cfg.Section("deppon").MapTo(DepponSetting)
}
