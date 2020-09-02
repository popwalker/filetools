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
	"path"
	"path/filepath"
	"time"

	"invtools/common"
	"invtools/pkg/pdfsplit"

	"invtools/utils"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	pdfsplitCmdExample = fmt.Sprintf("%s\n%s\n%s\n",
		fmt.Sprintf(`%s pdfsplit /input/directory`, appName),
		fmt.Sprintf(`%s pdfsplit /input/directory /output/directory`, appName),
		fmt.Sprintf(`%s pdfsplit /input/directory /output/directory -p=password -n=2`, appName),
	)
)

// pdfsplitCmd represents the pdfsplit command
var pdfsplitCmd = &cobra.Command{
	Use:   "pdfsplit",
	Short: "Split PDF files.",
	Long: `Split PDF files.

The command is used to extract one or more page ranges from pdf files.`,
	Example: pdfsplitCmdExample,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			inputPath, outputPath string
		)

		if len(args) < 1 {
			fmt.Println(aurora.Magenta("至少输入一个参数，比如pdf文件路径"))
			os.Exit(1)
		}
		inputPath = args[0]

		if len(args) == 2 {
			outputPath = args[1]
		} else {
			outputPath = path.Join(viper.GetString(common.CurrentDir), fmt.Sprintf("splited_pdfs_%s", time.Now().In(utils.LocationCST).Format("2006_01_02_15_04_05")))
		}

		if !path.IsAbs(inputPath) {
			if p, err := filepath.Abs(inputPath); err != nil {
				fmt.Println(aurora.Magenta("convert input directory to abs directory failed, please contact Rick~"))
				os.Exit(1)
			} else {
				inputPath = p
			}
		}

		if !path.IsAbs(outputPath) {
			if p, err := filepath.Abs(outputPath); err != nil {
				fmt.Println(aurora.Magenta("convert output directory to abs directory failed, please contact Rick~"))
				os.Exit(1)
			} else {
				outputPath = p
			}
		}

		//fmt.Println("[debug] input:", inputPath)
		//fmt.Println("[debug] output:", outputPath)
		//fmt.Println("[debug] password:", password)
		//fmt.Println("[debug] perpage:", perPage)
		//fmt.Println("[debug] concurrency:", concurrency)

		err := pdfsplit.NewPdfSplitter(inputPath, outputPath, password, perPage, concurrency).Do()
		if err != nil {
			fmt.Println(aurora.Magenta("拆分pdf出现错误，err:"), err)
			os.Exit(1)
		}

	},
}

var (
	password        string
	passwordFlag    = "password"
	perPage         int
	perPageFlag     = "perpage"
	concurrency     int
	concurrencyFlag = "split_concurrency"
)

func init() {
	rootCmd.AddCommand(pdfsplitCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pdfsplitCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pdfsplitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	pdfsplitCmd.Flags().StringVarP(&password, passwordFlag, "p", "", "PDF文件密码")
	pdfsplitCmd.Flags().IntVarP(&perPage, perPageFlag, "n", 1, "每N页做一次拆分")
	pdfsplitCmd.Flags().IntVarP(&concurrency, concurrencyFlag, "c", 1, "分N组并发拆分")
}
