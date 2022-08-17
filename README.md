# 简介
`rd` 是一个使用golang开发的研发工具, 集成了一系列常用的研发功能, 助力程序员提升研发效率.
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
rd search -t 搜索类型[打开默认浏览器/终端显示] -m 搜索引擎[bing/baidu/google/...] -s "搜索query" --desc=true
```

### 浏览器搜索
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


