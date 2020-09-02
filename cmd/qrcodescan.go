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
	"invtools/pkg/qrscan"

	"invtools/utils"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	qrcodescanCmdExample = fmt.Sprintf("%s\n%s\n",
		fmt.Sprintf(`%s qrcodescan /input/directory output.csv -t=qrcode`, appName),
		fmt.Sprintf(`%s qrcodescan /input/directory output.csv -c=2 -t=barcode128`, appName),
	)
)

// qrcodescanCmd represents the qrcodescan command
var qrcodescanCmd = &cobra.Command{
	Use:   "qrcodescan",
	Short: "Scan QRCode/Barcode",
	Long: `Scan QRCode/Barcode & get information.
support QRCode、Barcode128
`,
	Example: qrcodescanCmdExample,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			inputPath, outputPath string
		)

		if len(args) < 1 {
			fmt.Println(aurora.Magenta("至少输入一个参数，比如qrcode所在文件路径"))
			os.Exit(1)
		}
		inputPath = args[0]

		if len(args) == 2 {
			outputPath = args[1]
		} else {
			outputPath = path.Join(viper.GetString(common.CurrentDir), fmt.Sprintf("output_%s.csv", time.Now().In(utils.LocationCST).Format("20060102_15_04_05")))
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
		//fmt.Println("[debug] concurrency:", concurrency)
		//fmt.Println("[debug] qrType:", qrType)

		if err := qrscan.NewQrScanner(inputPath, outputPath, qrType, concurrency).Do(); err != nil {
			fmt.Println(aurora.Magenta("解析code出现错误，err:"), err)
			os.Exit(1)
		}
	},
}

var (
	qrConcurrencyFlag = "qr_concurrency"
	qrType            string
	qrTypeFlag        = "qr_type"
)

func init() {
	rootCmd.AddCommand(qrcodescanCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// qrcodescanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// qrcodescanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	qrcodescanCmd.Flags().IntVarP(&concurrency, qrConcurrencyFlag, "c", 1, "分N组并发解析")
	qrcodescanCmd.Flags().StringVarP(&qrType, qrTypeFlag, "t", "qrcode", "code类型,支持qrcode/barcode128")
}
