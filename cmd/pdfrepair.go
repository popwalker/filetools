// Copyright © 2020 NAME HERE <EMAIL ADDRESS>
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

	"invtools/pkg/pdfrepair"

	"invtools/utils"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var (
	pdfrepairCmdExample = fmt.Sprintf("%s\n%s\n",
		fmt.Sprintf(`%s pdfrepair /input/directory`, appName),
		fmt.Sprintf(`%s pdfrepair /input/directory /output/directory`, appName),
	)
)

// pdfrepairCmd represents the pdfrepair command
var pdfrepairCmd = &cobra.Command{
	Example: pdfrepairCmdExample,
	Use:     "pdfrepair",
	Short:   "Repair PDF files",
	Long:    `PDF修复,支持批量修复受损PDF文件`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			inputPath, outputPath string
		)

		if len(args) < 1 {
			fmt.Println(aurora.Magenta("至少输入一个参数，比如pdf文件所在目录。\n输入 invtools pdfrepaire -h 查看帮助"))
			os.Exit(1)
		}
		inputPath = args[0]

		if len(args) == 2 {
			outputPath = args[1]
		} else {
			outputPath = inputPath + "_repaired_" + time.Now().In(utils.LocationCST).Format("20060102150405")
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

		err := pdfrepair.NewPdfRepair(inputPath, outputPath).Do()
		if err != nil {
			fmt.Println(aurora.Magenta("修复pdf出现错误，err:"), err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pdfrepairCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pdfrepairCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pdfrepairCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
