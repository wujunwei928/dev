package cmd

import (
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// DefaultHttpPort flag默认端口
	DefaultHttpPort = 8899

	// HttpConfigPort http端口配置项
	HttpConfigPort = "http.port"

	// uploadMaxBytes 上传大小上限
	uploadMaxBytes = 100 << 20

	// serverReadHeaderTimeout 读请求头超时, 防御慢速连接(Slowloris)长期占用资源
	serverReadHeaderTimeout = 10 * time.Second

	// serverIdleTimeout 空闲连接超时
	serverIdleTimeout = 120 * time.Second
)

// httpRoutesOptions http服务路由配置
type httpRoutesOptions struct {
	StaticDir string // 静态文件根目录
	UseUpload bool   // 是否注册 /upload 上传端点
	Overwrite bool   // 上传是否允许覆盖已存在文件
}

func NewCmdHttp() *cobra.Command {
	var (
		useUpload bool
		overwrite bool
		lanExpose = true
	)

	cmd := &cobra.Command{
		Use:   "http",
		Short: "http服务",
		Long:  `启动http服务上传下载文件, 类似python的http.server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			strHttpPort := strconv.Itoa(viper.GetInt(HttpConfigPort))

			staticDir := "./"
			if pwd, err := os.Getwd(); err == nil {
				staticDir = pwd
			}

			// 独立mux, 不污染全局DefaultServeMux
			mux := http.NewServeMux()
			registerHttpRoutes(mux, httpRoutesOptions{
				StaticDir: staticDir,
				UseUpload: useUpload,
				Overwrite: overwrite,
			})

			// 默认监听局域网方便跨设备传输; --lan=false 时仅监听本机回环
			listenAddr := "127.0.0.1:" + strHttpPort
			if lanExpose {
				listenAddr = ":" + strHttpPort
			}

			srv := &http.Server{
				Addr:              listenAddr,
				Handler:           mux,
				ReadHeaderTimeout: serverReadHeaderTimeout,
				IdleTimeout:       serverIdleTimeout,
			}

			if lanExpose {
				// 打印静态文件服务器地址, 方便访问
				localIp, _ := GetLocalIp() // 获取本机IP
				fmt.Println("start listen: http://" + localIp + ":" + strHttpPort)
			} else {
				fmt.Println("start listen: http://127.0.0.1:" + strHttpPort)
			}
			return srv.ListenAndServe()
		},
	}

	// viper配置 和 flag 绑定, 使用 viper.Get 获取对应配置项时的顺序, 因为后续需要用viper获取flag, 这里不需要用IntVarP, 用IntP即可
	// 参考: https://github.com/spf13/cobra/issues/23
	// If no config file is specified, use default flag value.
	// If config file is specified, and the user did not set the flag, use the value in the config file
	// If config file is specified, but the user did set the flag, use the value specified by the user, even if it's just the default flag value.
	cmd.Flags().IntP("port", "p", 8899, "http端口")
	_ = viper.BindPFlag(HttpConfigPort, cmd.Flags().Lookup("port"))

	// 上传端点默认关闭, 需显式开启
	cmd.Flags().BoolVarP(&useUpload, "use_upload", "", false, "启用上传功能(注册/upload端点), 默认关闭")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "", false, "上传时允许覆盖已存在文件, 默认拒绝覆盖(需与--use_upload同时使用)")
	cmd.Flags().BoolVarP(&lanExpose, "lan", "", true, "监听局域网(0.0.0.0)方便跨设备传输, false时仅监听127.0.0.1")

	return cmd
}

// registerHttpRoutes 注册静态下载/首页/上传路由
func registerHttpRoutes(mux *http.ServeMux, opts httpRoutesOptions) {
	// 静态文件, 文件下载
	fs := http.FileServer(http.Dir(opts.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// 文件上传页面
	mux.HandleFunc("/", newIndexHandler(opts.UseUpload))

	// 上传端点仅在显式启用时注册, 未启用时不暴露任何写入面
	if opts.UseUpload {
		mux.HandleFunc("/upload", newUploadHandler(opts.StaticDir, opts.Overwrite))
	}
}

// newIndexHandler 首页handler, 仅响应根路径
func newIndexHandler(useUpload bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		tmpl := template.Must(template.New("http upload").Parse(indexPageTemplate))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(w, map[string]any{
			"use_upload": useUpload,
		})
	}
}

// newUploadHandler 上传handler: 文件名取basename防路径穿越, 默认拒绝覆盖已存在文件
func newUploadHandler(saveDir string, overwrite bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 限制上传大小为100MB
		r.Body = http.MaxBytesReader(w, r.Body, uploadMaxBytes)

		f, fh, err := r.FormFile("upload-file")
		if err != nil {
			http.Error(w, "FormFile fail:"+err.Error(), http.StatusBadRequest)
			return
		}
		defer f.Close()

		// 使用 filepath.Base 防止路径穿越
		safeFilename := filepath.Base(fh.Filename)
		fileFullSavePath := filepath.Join(saveDir, safeFilename)

		// 拒绝覆盖已存在文件, 防止局域网内静默替换本地文件
		if !overwrite {
			if _, err := os.Stat(fileFullSavePath); err == nil {
				http.Error(w, "file already exists (use --overwrite to replace): "+fileFullSavePath, http.StatusConflict)
				return
			}
		}

		// #nosec G304 -- 文件名已经过 filepath.Base 中和路径穿越
		dst, err := os.Create(fileFullSavePath)
		if err != nil {
			http.Error(w, "Create file fail:"+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, f); err != nil {
			http.Error(w, "Copy file fail:"+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// #nosec G705 -- text/plain 响应且客户端以 textContent 渲染, 无 HTML 执行上下文
		fmt.Fprintf(w, "file upload success: %s", fileFullSavePath)
	}
}

// indexPageTemplate 首页模板: 原生fetch上传, 结果以textContent渲染, 不引入外部JS
const indexPageTemplate = `
<html>
    <head>
        <title>Upload File</title>
        <style>
         .drag-wrapper{
             width:100%;
             height: 300px;
             border: 4px dashed lightblue;
             text-align: center;
             line-height: 300px;
             color: lightgrey;
             font-size: 36px;
         }
        </style>
    </head>
    <body>

    <h1 style="margin: 36px;"><a href="/static">下载服务器文件</a></h1>

    <hr/>
	{{if .use_upload}}
    <h1>上传文件</h1>
    <form >
        <div class="drag-wrapper" id="drag-wrapper">
            请将文件拖到此处, 自动上传
        </div>
    </form>
        <script>
         (function () {
             var drop = document.getElementById('drag-wrapper');
             drop.addEventListener('dragover', function (event) {
                 event.preventDefault();
             });
             drop.addEventListener('drop', function (event) {
                 event.preventDefault();
                 var file = event.dataTransfer.files[0];
                 if (!file) {
                     return;
                 }
                 var formData = new FormData();
                 formData.append('upload-file', file);
                 fetch('/upload', {method: 'POST', body: formData})
                     .then(function (response) {
                         return response.text().then(function (text) {
                             return {ok: response.ok, text: text};
                         });
                     })
                     .then(function (result) {
                         showResult(result.text);
                     })
                     .catch(function (err) {
                         showResult('upload fail: ' + err);
                     });
             });
             // 结果以textContent渲染, 防止文件名/错误信息被注入为HTML
             function showResult(text) {
                 var div = document.createElement('div');
                 div.style.marginTop = '20px';
                 div.style.fontSize = '28px';
                 div.textContent = text;
                 drop.after(div);
             }
         })();
        </script>
	{{end}}
    </body>
</html>`

func GetLocalIp() (string, error) {
	conn, err := net.Dial("udp", "google.com:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}
