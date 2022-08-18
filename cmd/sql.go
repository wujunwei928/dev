package cmd

import (
	"log"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
	"github.com/wujunwei928/dev/internal/sql2struct"
)

// sql默认项
const (
	DefaultSqlType    = "mysql"          // 默认搜索引擎
	DefaultSqlHost    = "127.0.0.1:3306" // 默认检索类型
	DefaultSqlCharset = "utf8mb4"        // 默认cli是否倒序显示
)

// sql配置项
const (
	SqlConfigType     = "sql.type"
	SqlConfigHost     = "sql.host"
	SqlConfigUserName = "sql.username"
	SqlConfigPassword = "sql.password"
	SqlConfigDb       = "sql.db"
	SqlConfigCharset  = "sql.charset"
)

var tableName string

// sqlCmd represents the sql command
var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "sql转换和处理",
	Long:  "sql转换和处理",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var sql2structCmd = &cobra.Command{
	Use:   "struct",
	Short: "sql转换",
	Long:  "sql转换",
	Run: func(cmd *cobra.Command, args []string) {
		username := viper.GetString(SqlConfigUserName)
		password := viper.GetString(SqlConfigPassword)
		host := viper.GetString(SqlConfigHost)
		charset := viper.GetString(SqlConfigCharset)
		dbType := viper.GetString(SqlConfigType)
		dbName := viper.GetString(SqlConfigDb)

		dbInfo := &sql2struct.DBInfo{
			DBType:   dbType,
			Host:     host,
			UserName: username,
			Password: password,
			Charset:  charset,
		}
		dbModel := sql2struct.NewDBModel(dbInfo)
		err := dbModel.Connect()
		if err != nil {
			log.Fatalf("dbModel.Connect err: %v", err)
		}
		columns, err := dbModel.GetColumns(dbName, tableName)
		if err != nil {
			log.Fatalf("dbModel.GetColumns err: %v", err)
		}

		template := sql2struct.NewStructTemplate()
		templateColumns := template.AssemblyColumns(columns)
		err = template.Generate(tableName, templateColumns)
		if err != nil {
			log.Fatalf("template.Generate err: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(sqlCmd)

	sqlCmd.AddCommand(sql2structCmd)
	sql2structCmd.Flags().StringP("username", "", "", "请输入数据库的账号")
	sql2structCmd.Flags().StringP("password", "", "", "请输入数据库的密码")
	sql2structCmd.Flags().StringP("host", "", DefaultSqlHost, "请输入数据库的HOST")
	sql2structCmd.Flags().StringP("charset", "", DefaultSqlCharset, "请输入数据库的编码")
	sql2structCmd.Flags().StringP("type", "", DefaultSqlType, "请输入数据库实例类型")
	sql2structCmd.Flags().StringP("db", "", "", "请输入数据库名称")
	sql2structCmd.Flags().StringVarP(&tableName, "table", "", "", "请输入表名称")

	// flag 和 viper 绑定
	viper.BindPFlag(SqlConfigUserName, sql2structCmd.Flags().Lookup("username"))
	viper.BindPFlag(SqlConfigPassword, sql2structCmd.Flags().Lookup("password"))
	viper.BindPFlag(SqlConfigHost, sql2structCmd.Flags().Lookup("host"))
	viper.BindPFlag(SqlConfigCharset, sql2structCmd.Flags().Lookup("charset"))
	viper.BindPFlag(SqlConfigType, sql2structCmd.Flags().Lookup("type"))
	viper.BindPFlag(SqlConfigDb, sql2structCmd.Flags().Lookup("db"))
}
