   * [简介](#简介)
   * [安装](#安装)
      * [go install 安装](#go-install-安装)
      * [下载可执行文件](#下载可执行文件)
   * [配置](#配置)
   * [命令列表](#命令列表)
      * [search: 搜索服务](#search-搜索服务)
         * [浏览器搜索](#浏览器搜索)
         * [命令行](#命令行)
         * [设置默认配置, 减少命令行书写](#设置默认配置-减少命令行书写)
      * [open: 打开网址或文件夹](#open-打开网址或文件夹)
      * [http: http服务](#http-http服务)
      * [encode: 字符串加密](#encode-字符串加密)
      * [decode: 字符串解密](#decode-字符串解密)
      * [time: 时间转换](#time-时间转换)
      * [json: json工具](#json-json工具)
      * [sql: sql工具](#sql-sql工具)
      * [word: 单词工具](#word-单词工具)
   * [依赖模块](#依赖模块)
   * [常见问题](#常见问题)

# 简介
`rd` 是一个使用golang开发的研发工具, 集成了一系列常用的研发功能, 助力程序员提升研发效率.
使用go开发是可以方便的编译为二进制, 没有脚本语言的包依赖问题.

集成了以下功能:

|功能|说明|
|---|---|
| search | 搜索: 支持打开默认浏览器搜索 和 终端显示搜索结果 |
| open | 打开网址或文件夹 |
| http | http服务: 在运行文件夹启动http服务, 支持下载文件和上传文件, 方便文件传输 |
| encode | 字符串加密: md5, sha1, base64, url, unicode... |
| decode | 字符串解密: base64, url, unicode |
| time | 时间戳转时间, 时间转时间戳 |
| json | json转golang结构体 |
| sql | sql转golang结构体 |
| word | 下划线转驼峰, 驼峰转下划线 |

**查看帮助**
```bash
rd -h
```
```bash
研发工具箱

Usage:
  rd [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  decode      字符串解密
  encode      字符串加密
  help        Help about any command
  http        http服务
  json        json转换和处理
  open        打开网址或文件路径
  search      搜索
  sql         sql转换和处理
  time        时间格式处理
  word        单词格式转换

Flags:
      --config string   config file (default is $HOME/.rd.yaml)
  -h, --help            help for rd

Use "rd [command] --help" for more information about a command.
```

# 安装
## go install 安装
如果你本机安装有go sdk, 可以直接用go install 安装最新版
```bash
go install github.com/wujunwei928/rd@latest
```

## 下载可执行文件
下载指定系统的可执行文件  
下载地址: https://github.com/wujunwei928/rd/releases

# 配置
工具支持自定义配置, 检索相关命令
默认配置路径: `$HOME/.rd.yaml`

支持配置:
```yaml
http:
# http配置
    port: 8811  # http端口
search:
# 检索配置
    cli_is_desc: true       # 终端显示结果是否倒序
    default_engine: kaifa   # 默认搜索引擎
    default_type: cli       # 默认检索模式: browser:打开默认浏览器检索; cli: 终端显示搜索结果
sql:
# sql配置
    type: mysql             # 数据库类型
    host: 127.0.0.1:3306    # 数据库host
    username: root          # 数据库用户名
    password: 123456        # 数据库密码
    db: novel               # 数据库名称
    charset: utf8mb4        # 字符类型
```

flag默认值, flag配置文件配置项, flag用户手动设置项 优先级说明: 
1. 如果配置文件没有设置对应配置项, 使用flag的默认值
2. 如果配置文件有设置配置项, 用户没有设置flag值, 使用配置文件配置项
3. 如果配置文件有设置配置项, 但是用户设置了flag值, 使用用户设置的值, 即使设置的值和默认值相等, 

# 命令列表
## search: 搜索服务
**查看帮助**
```bash
$ rd search -h
指定搜索引擎, 检索相关query

Usage:
  rd search [flags]

Flags:
      --desc          是否倒序展示: 默认倒序, 方便查看(只终端展示生效) (default true)
  -h, --help          help for search
  -m, --mode string   打开默认浏览器, 指定搜索引擎, 检索相关query，模式如下：
                      360: 360搜索
                      sogou: 搜狗搜索
                      baidu: 百度搜索
                      zhihu: 知乎搜索
                      weixin: 搜狗微信搜索
                      kaifa: 百度开发者搜索
                      douban: 豆瓣搜索
                      movie: 豆瓣电影搜索
                      book: 豆瓣书籍搜索
                      bing: 必应搜索
                      google: 谷歌搜索
                      github: Github搜索
  -s, --str string    请输入搜索query
  -t, --type string   检索方式:
                      browser: 打开默认浏览器检索
                      cli: 终端显示搜索内容

Global Flags:
      --config string   config file (default is $HOME/.rd.yaml)

```



**使用方式**
```bash
# 长标签模式
rd search --type=搜索类型[打开默认浏览器/终端显示] --mode=搜索引擎[bing/baidu/google/...] --str="搜索query" --desc=true

# 短标签模式
rd search -t 搜索类型[打开默认浏览器/终端显示] -m 搜索引擎[bing/baidu/google/...] -s "搜索query" --desc=true
```

### 浏览器搜索
**指定搜索引擎, 检索query**
```bash
rd search -m bing -s "golang slice"
```

**不指定query, 浏览器打开搜索引擎首页**
```bash
rd search -m bing
```

### 命令行
在终端里直接显示检索结果, 适用于习惯命令行或服务器没有浏览器的模式下使用, 因为结果显示信息较多, 为了方便查看, `搜索结果默认倒序显示`
> 1. 命令行模式必须指定query
> 2. windows下尽量使用windows terminal, 支持鼠标点击链接跳转

**默认倒序显示**
```bash
rd search -t cli -m bing -s "golang slice" --desc=false
```

**强制正序显示**
```bash
rd search -t cli -m bing -s "golang slice" --desc=false
```

### 设置默认配置, 减少命令行书写
```yaml
search:
  default-type: cli
  default-engine: bing
```

## open: 打开网址或文件夹
**打开文件夹**
```bash
rd open -s .
```

**使用默认浏览器打开网址**
```bash
rd open -s https://www.baidu.com/
```
> ps: 网址必须带协议

## http: http服务
快速启动http服务, 方便文件传输, 支持文件上传下载
默认端口: 8899, 支持命令行或配置自定义

**启动http服务**
```bash
rd http
```

**指定端口启动**
```bash
rd http -p 8080
```

**主界面**

![image](https://user-images.githubusercontent.com/3396697/185265872-2bf24b42-1281-442e-8cf6-5eb90e4f93ac.png)


**下载界面**

![image](https://user-images.githubusercontent.com/3396697/185266236-d714b180-47ae-40ba-a62e-6bd6bbaba3b9.png)


## encode: 字符串加密
字符串加密: 支持 md5, sha1, base64, url加密, unicode...

**使用方式**
```bash
# md5加密
rd encode -m md5 -s golang

# sha1加密
rd encode -m sha1 -s golang

# base64加密
rd encode -m base64 -s golang

# url加密
rd encode -m url -s "name=张三&age=18"

# sha1加密
rd encode -m unicode -s 中国人
```

## decode: 字符串解密
字符串解密, 支持: base64, url, unicode...

**打开文件夹**
```bash
# base64
rd decode -m base64 -s 5Lit5Zu95Lq6

# url
rd decode -m url -s name%3D%E5%BC%A0%E4%B8%89%26age%3D18

# unicode
rd decode -m unicode -s "\u4e2d\u56fd\u4eba"
```

## time: 时间转换
**解析时间戳**
```bash
# 解析当前时间戳
rd time parse

# 解析指定时间戳
rd time parse -t 123
```

**计算时间**
```bash
# 获取某个时间的时间戳
rd time calc -c "2022-08-17 19:40:11" -d 0

# 指定时间增加10分钟 (支持 "ns", "us" (or "µ s"), "ms", "s", "m", "h")
rd time calc -c "2022-08-17 19:40:11" -d +10m

# 指定时间减少10分钟
rd time calc -c "2022-08-17 19:40:11" -d -10m
```

## json: json工具
**json转golang结构体**
```bash
rd json struct -s '{"name":"zhangsan","list":["a", "b", "c"]}'
```

## sql: sql工具
**转golang结构体**
**查看帮助**
```bash
rd sql struct -h
```
```bash
sql转换

Usage:
  rd sql struct [flags]

Flags:
      --charset string    请输入数据库的编码 (default "utf8mb4")
      --db string         请输入数据库名称
  -h, --help              help for struct
      --host string       请输入数据库的HOST (default "127.0.0.1:3306")
      --password string   请输入数据库的密码
      --table string      请输入表名称
      --type string       请输入数据库实例类型 (default "mysql")
      --username string   请输入数据库的账号

Global Flags:
      --config string   config file (default is $HOME/.rd.yaml)
```

**指定数据表生成结构体**
```bash
# rd sql struct --type=数据库类型 --host=数据库host --username=用户名 --password=密码 --db=数据库名 --table=表名
rd sql struct --type=mysql --host="127.0.0.1:3306" --username=root --password=123456 --db=blog --table=user
```

## word: 单词工具
**查看帮助**
```bash
rd word -h
```

**使用方式**
```bash
# 转大写
rd word -m 1 -s abc

# 转小写
rd word -m 2 -s ABC

# 下划线转驼峰
rd word -m 3 -s abc_def

# 下划线转驼峰(首个单词首字母小写)
rd word -m 4 -s abc_def

# 驼峰转下划线
rd word -m 5 -s AbcDefGhk
```


# 依赖模块
|模块|作用|
|---|---|
| [github.com/spf13/cobra](https://github.com/spf13/cobra) | 命令行交互 |
| [github.com/spf13/viper](https://github.com/spf13/viper) | 配置管理 |
| [github.com/go-resty/resty/v2](https://github.com/go-resty/resty/v2) | HTTP 和 REST 客户端 |
| [github.com/tidwall/gjson](https://github.com/tidwall/gjson) | 使用一行代码获取JSON的值 |
| [github.com/PuerkitoBio/goquery](https://github.com/PuerkitoBio/goquery)  | jQuery语法解析html页面 |
| [github.com/mitchellh/go-homedir](https://github.com/mitchellh/go-homedir)  | 用于检测用户的主目录 |


# 常见问题
1. 如果相关参数值值有空格或特殊符号(如:&), 需要用双引号 "s=golang slice&b=2"

