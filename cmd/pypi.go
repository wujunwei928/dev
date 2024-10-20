package cmd

import (
	"os/exec"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(NewCmdPypi())
}

type pypiSubCmd struct {
	mirrorList []string
}

func NewCmdPypi() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pypi",
		Short: "设置pypi镜像",
		Long:  `设置pypi镜像`,
	}

	// 子命令
	subCmd := pypiSubCmd{
		mirrorList: []string{
			"https://mirrors.tuna.tsinghua.edu.cn/pypi/web/simple",
			"https://mirrors.163.com/pypi/simple/",
		},
	}
	cmd.AddCommand(subCmd.NewCmdLs())
	cmd.AddCommand(subCmd.NewCmdUse())

	return cmd
}

func (p *pypiSubCmd) getUsePypi() (string, error) {
	// 获取当前使用的pypi源: pip config get global.index-url
	getUsePypiCmd := exec.Command("pip", "config", "get", "global.index-url")
	getUsePypi, err := getUsePypiCmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(getUsePypi)), nil
}

func (p *pypiSubCmd) setUsePypi(pypiUrl string) (string, error) {
	// 设置当前使用的pypi源: pip config set global.index-url https://xxx
	cmd := exec.Command("pip", "config", "set", "global.index-url", pypiUrl)
	usePypi, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(usePypi)), nil
}

func (p *pypiSubCmd) NewCmdLs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List all the registries",
		Long:  `List all the registries`,
		RunE: func(cmd *cobra.Command, args []string) error {
			getUsePypi, err := p.getUsePypi()
			if err != nil {
				return err
			}

			bulletListItems := make([]pterm.BulletListItem, 0, len(p.mirrorList))
			for _, s := range p.mirrorList {
				bullet := " "
				if s == getUsePypi {
					bullet = pterm.Green("*")
				}

				bulletListItems = append(bulletListItems, pterm.BulletListItem{
					Level:  2,
					Text:   s,
					Bullet: bullet,
				})
			}

			pterm.BulletListPrinter{
				TextStyle:   &pterm.ThemeDefault.BulletListTextStyle,
				BulletStyle: &pterm.ThemeDefault.BulletListBulletStyle,
			}.WithItems(bulletListItems).Render()
			return nil
		},
	}

	return cmd
}

func (p *pypiSubCmd) NewCmdUse() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Change current registry",
		Long:  `Change current registry`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var options []string
			for _, s := range p.mirrorList {
				options = append(options, s)
			}

			selectedOption, _ := pterm.DefaultInteractiveSelect.WithOptions(options).Show()

			// Display the selected option to the user with a green color for emphasis
			pterm.Info.Printfln("use pypi: %s", pterm.Green(selectedOption))

			setRes, err := p.setUsePypi(selectedOption)
			if err != nil {
				return err
			}
			pterm.Info.Printfln(setRes)

			return nil
		},
	}

	return cmd
}
