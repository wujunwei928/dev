package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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

type sqlSubCmd struct{}

func NewCmdSql() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sql",
		Short: "sql转换和处理",
		Long:  "sql转换和处理",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	// 添加子命令
	subCmd := sqlSubCmd{}
	cmd.AddCommand(subCmd.NewCmdStruct())

	return cmd
}

func (s sqlSubCmd) NewCmdStruct() *cobra.Command {
	var (
		tableName string
	)

	cmd := &cobra.Command{
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

	cmd.Flags().StringP("username", "", "", "请输入数据库的账号")
	cmd.Flags().StringP("password", "", "", "请输入数据库的密码")
	cmd.Flags().StringP("host", "", DefaultSqlHost, "请输入数据库的HOST")
	cmd.Flags().StringP("charset", "", DefaultSqlCharset, "请输入数据库的编码")
	cmd.Flags().StringP("type", "", DefaultSqlType, "请输入数据库实例类型")
	cmd.Flags().StringP("db", "", "", "请输入数据库名称")
	cmd.Flags().StringVarP(&tableName, "table", "", "", "请输入表名称")

	// flag 和 viper 绑定
	viper.BindPFlag(SqlConfigUserName, cmd.Flags().Lookup("username"))
	viper.BindPFlag(SqlConfigPassword, cmd.Flags().Lookup("password"))
	viper.BindPFlag(SqlConfigHost, cmd.Flags().Lookup("host"))
	viper.BindPFlag(SqlConfigCharset, cmd.Flags().Lookup("charset"))
	viper.BindPFlag(SqlConfigType, cmd.Flags().Lookup("type"))
	viper.BindPFlag(SqlConfigDb, cmd.Flags().Lookup("db"))

	return cmd
}
