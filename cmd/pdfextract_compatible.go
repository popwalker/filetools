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

	"invtools/common"
	"invtools/pkg/pdfextract/compatible"

	. "github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var compatibleCmdExample = fmt.Sprintf("%s\n",
	fmt.Sprintf(`%s pdfextract compatible /input/directory output.csv coord_name="coord1 coord2 coord3 coord4" reg_name="regexp" --with_coordinate=true --with_ocr=true`, appName),
)

// compatibleCmd represents the compatible command
var compatibleCmd = &cobra.Command{
	Use:     "compatible",
	Short:   "兼容模式解析PDF",
	Example: compatibleCmdExample,
	Long:    `兼容模式，允许使用坐标和OCR的方式同时解析PDF文件`,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("compatible called")
		if len(args) < 2 {
			fmt.Println(Magenta("Please input 3 arguments at least."))
			//cmd.Help();
			os.Exit(1)
		}

		var (
			inputDir            = args[0]
			outputFile          = args[1]
			parsedArgs []string = args[2:]
		)

		if inputDir == "." || inputDir == "./" {
			inputDir = common.CurrentDir
		}
		if !path.IsAbs(inputDir) {
			p, err := filepath.Abs(inputDir)
			if err != nil {
				fmt.Println(Magenta("convert inputDir to abs failed"))
				os.Exit(1)
			}
			inputDir = p
		}

		if !path.IsAbs(outputFile) {
			p, err := filepath.Abs(outputFile)
			if err != nil {
				fmt.Println(Magenta("convert outputFile to abs failed"))
				os.Exit(1)
			}
			outputFile = p
		}

		//fmt.Println("[debug] input:", inputDir)
		//fmt.Println("[debug] output:", outputFile)
		//fmt.Println("[debug] parsedArgs:", parsedArgs, ",", len(parsedArgs))
		//fmt.Println("[debug] concurrency:", compatibleConcurrency)
		//fmt.Println("[debug] with-coordinate:", withCoordinate)
		//fmt.Println("[debug] with-ocr:", withOcr)
		//fmt.Println("[debug] with-conf:", cnf)

		err := compatible.NewExtractor(
			inputDir,
			outputFile,
			compatibleConcurrency,
			maxReadPage,
			withCoordinate,
			withOcr,
			parsedArgs,
			cnf,
			debug,
		).Extract()
		if err != nil {
			fmt.Println(Magenta(fmt.Sprintf("Extract from pdf voucher failed, inputDir: %s ,err:%+v",inputDir, err)))
		}
	},
}

var (
	// 并发执行的flag名称
	compatibleConcurrency     int
	compatibleConcurrencyFlag = "compatible_concurrency"

	// 是否使用坐标
	withCoordinate     bool
	withCoordinateFlag = "with_coordinate"

	// 是否使用OCR
	withOcr     bool
	withOcrFlag = "with_ocr"

	maxReadPage     int
	maxReadPageFlag = "max_read_page"

	cnf string
	cnfFlag = "with_cnf"

	debug bool
	debugFlag = "with_debug"
)

func init() {
	pdfextractCmd.AddCommand(compatibleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compatibleCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compatibleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	compatibleCmd.Flags().IntVarP(&compatibleConcurrency, compatibleConcurrencyFlag, "n", 1, "分N组并发解析")
	//viper.BindPFlag(compatibleConcurrencyFlag, coordinateCmd.Flags().Lookup(compatibleConcurrencyFlag))

	compatibleCmd.Flags().BoolVar(&withCoordinate, withCoordinateFlag, false, "使用坐标解析(default false)")
	//coordinateCmd.Flags().BoolVarP(&withCoordinate, withCoordinateFlag, "c", false, "使用坐标解析")

	//coordinateCmd.Flags().BoolVarP(&withOcr, withOcrFlag, "o", false, "使用OCR解析")
	compatibleCmd.Flags().BoolVar(&withOcr, withOcrFlag, false, "使用OCR解析(default false)")

	compatibleCmd.Flags().IntVarP(&maxReadPage, maxReadPageFlag, "m", 0, "每张pdf最多读取页数(default 0)")

	compatibleCmd.Flags().StringVarP(&cnf, cnfFlag, "c", "", "配置文件路径")

	compatibleCmd.Flags().BoolVarP(&debug, debugFlag, "d", false, "是否开启debug")
}
