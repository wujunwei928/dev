package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

const (
	// DefaultHttpPort flag默认端口
	DefaultHttpPort = 8899

	// HttpConfigPort http端口配置项
	HttpConfigPort = "http.port"
)

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "http服务",
	Long:  `启动http服务上传下载文件, 类似python的http.server`,
	Run: func(cmd *cobra.Command, args []string) {
		strHttpPort := strconv.Itoa(viper.GetInt(HttpConfigPort))

		// 静态文件, 文件下载
		fs := http.FileServer(http.Dir("./"))
		http.Handle("/static/", http.StripPrefix("/static/", fs))

		// 文件上传页面
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

			tmpl := template.Must(template.New("http upload").Parse(`
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
        .drag-wrapper img{
            max-width: 80px;
        }
        </style>
    </head>
    <body>

    <h1 style="margin: 36px;"><a href="/static">下载服务器文件</a></h1>

    <hr/>

    <h1>上传文件</h1>
    <form >
        <div class="drag-wrapper">
            请将文件拖到此处, 自动上传
        </div>
    </form>
        <script src="https://code.jquery.com/jquery-2.2.3.min.js"></script>
        <script>
         $('.drag-wrapper').on('dragover',function(event){
              event.preventDefault()
         }).on('drop',function(event){

             event.preventDefault();
             //数据在event的dataTransfer对象里
             let file = event.originalEvent.dataTransfer.files[0];
             //同样用fileReader实现图片上传
             let fd = new FileReader();
             let fileType = file.type;
             fd.readAsDataURL(file);
             fd.onload = function(){
                  // if(/^image\/[jpeg|png|gif|jpg]/.test(fileType)){
                  //     $('.drag-wrapper').append('<img src="'+this.result+'"/>')
                  // }else{
                  //     alert('仅支持拖拽图片')
                  // }
             }
             let formData = new FormData();
             formData.append('upload-file',file);
             $.ajax({
                url : "/upload",
                type : 'POST',
                cache : false,
                data : formData,
                processData : false,
                contentType : false,
                success : function(result) {
                    // do something
                    tips = $('<div></div>').text(result)
                    $('.drag-wrapper').after("<div style='margin-top: 20px; font-size: 28px'>"+ result +"<div>")
                }
            });
         })
        </script>
    </body>
</html>`))
			tmpl.Execute(w, nil)
		})

		// 上传文件
		http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
			f, fh, err := r.FormFile("upload-file")
			if err != nil {
				//t.Fatalf("FormFile(%q): %q", key, err)
				fmt.Fprintf(w, "FormFile fail:"+err.Error())
			}
			var b bytes.Buffer
			_, err = io.Copy(&b, f)
			if err != nil {
				//t.Fatal("copying contents:", err)
				fmt.Fprintf(w, "copying contents:"+err.Error())
			}
			err = ioutil.WriteFile(fh.Filename, b.Bytes(), 0644)
			if err != nil {
				fmt.Fprintf(w, "write upload file: "+err.Error())
			}

			// 拼接上传文件保存绝对路径
			pwd, _ := os.Getwd()
			fileFullSavePath := filepath.Join(pwd, fh.Filename)
			fmt.Fprintf(w, "file upload success: "+fileFullSavePath)
		})

		// 打印静态文件服务器地址, 方便访问
		localIp, _ := GetLocalIp() // 获取本机IP
		fmt.Println("start listen: http://" + localIp + ":" + strHttpPort)
		http.ListenAndServe(":"+strHttpPort, nil)

	},
}

func init() {
	rootCmd.AddCommand(httpCmd)

	// viper配置 和 flag 绑定, 使用 viper.Get 获取对应配置项时的顺序, 因为后续需要用viper获取flag, 这里不需要用IntVarP, 用IntP即可
	// 参考: https://github.com/spf13/cobra/issues/23
	// If no config file is specified, use default flag value.
	// If config file is specified, and the user did not set the flag, use the value in the config file
	// If config file is specified, but the user did set the flag, use the value specified by the user, even if it's just the default flag value.
	httpCmd.Flags().IntP("port", "p", 8899, "http端口")
	viper.BindPFlag(HttpConfigPort, httpCmd.Flags().Lookup("port"))
}

func GetLocalIp() (string, error) {
	conn, err := net.Dial("udp", "google.com:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0], nil
}
