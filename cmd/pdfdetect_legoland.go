// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"invtools/common"
	"invtools/pkg/detective"
	"invtools/pkg/detective/legoland"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var legolandCmdExample = fmt.Sprintf("%s\n%s\n%s\n",
	fmt.Sprintf(`%s pdfdetect legoland -d="/path/to/your/directory" -a="expected act name" -v="expected effective date"`, appName),
	fmt.Sprintf(`%s pdfdetect legoland -d="/path/to/your/directory" -a="expected act name" -v="expected effective date" -c`, appName),
	fmt.Sprintf(`%s pdfdetect legoland "/path/to/your/directory" "expected act name" "expected effective date"`, appName),
)

// legolandCmd represents the legoland command
var legolandCmd = &cobra.Command{
	Use:     "legoland",
	Short:   "parse & analyze legoland pdf.",
	Example: legolandCmdExample,
	Long: `
检测legoland票券信息.`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("legoland called")
		var (
			dir      = viper.GetString(common.LeGolandFlagDir)
			activity = viper.GetString(common.LeGolandFlagActivity)
			valid    = viper.GetString(common.LeGolandFlagValid)
		)
		if dir == "" && args[0] != "" {
			dir = args[0]
		}

		if activity == "" && args[1] != "" {
			activity = args[1]
		}

		if valid == "" && args[2] != "" {
			valid = args[2]
		}

		// 执行Legoland解析
		detectiveLeGoland := legoland.NewDetective(dir, activity, valid)
		err := detective.Detect(detectiveLeGoland)
		if err != nil {
			fmt.Println("Legoland detect failed. err:", err)
		}
	},
	Args: func(cmd *cobra.Command, args []string) error {
		var (
			dir      = viper.GetString(common.LeGolandFlagDir)
			activity = viper.GetString(common.LeGolandFlagActivity)
			valid    = viper.GetString(common.LeGolandFlagValid)
		)
		if len(args) == 0 {
			if dir == "" {
				viper.Set(common.LeGolandFlagDir, common.CurrentDir)
				dir = common.CurrentDir
			}
			if dir == "" || activity == "" || valid == "" {
				cmd.Help()
				os.Exit(1)
			}
		} else if len(args) < 2 {
			cmd.Help()
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	pdfdetectCmd.AddCommand(legolandCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// legolandCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// legolandCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// 保存pdf的目录
	legolandCmd.Flags().StringP("dir", "d", "", "The directory where pdf is.")
	viper.BindPFlag(common.LeGolandFlagDir, legolandCmd.Flags().Lookup(common.LeGolandFlagDir))

	// 待匹配的活动名称
	legolandCmd.Flags().StringP("activity", "a", "", "Activity name which need to be detected.")
	viper.BindPFlag(common.LeGolandFlagActivity, legolandCmd.Flags().Lookup(common.LeGolandFlagActivity))

	// 待匹配的有效期
	legolandCmd.Flags().StringP("valid", "v", "", "Date which need to be detected.")
	viper.BindPFlag(common.LeGolandFlagValid, legolandCmd.Flags().Lookup(common.LeGolandFlagValid))

	legolandCmd.Flags().BoolP("classify", "c", false, "Whether classify unexpected file together")
	viper.BindPFlag(common.LeGolandFlagClassify, legolandCmd.Flags().Lookup(common.LeGolandFlagClassify))

	viper.Set(common.RunningDetective, legoland.DetectiveNickName)
}
