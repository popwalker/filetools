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

	"invtools/common"
	"invtools/pkg/pdfextract"
	"invtools/pkg/pdfextract/coordinate"

	. "github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var coordinateCmdExample = fmt.Sprintf("%s\n%s\n%s\n",
	fmt.Sprintf(`%s pdfextract coordinate -i="/input/directory" -o=output.csv`, appName),
	fmt.Sprintf(`%s pdfextract coordinate -i="/input/directory" -o=output.xlsx -c=2`, appName),
	fmt.Sprintf(`%s pdfextract coordinate /input/directory output.csv position="coord1 coord2 coord3 coord4"`, appName),
)

// coordinateCmd represents the common command
var coordinateCmd = &cobra.Command{
	Use:   "coordinate",
	Short: "Extract information from pdf via coordination",
	Long: `Extract information from pdf via coordination
Only support single page pdf for now.
`,
	Example: coordinateCmdExample,
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("coordinate called")

		if len(args) < 3 {
			fmt.Println(Magenta("Please input 3 arguments at least."))
			//cmd.Help();
			os.Exit(1)
		}

		var (
			input       = args[0]
			output      = args[1]
			coordinates []string
			concurrency = viper.GetInt(common.CoordinateExtractFlagConcurrency)
		)

		if input == "." || input == "./" {
			input = common.CurrentDir
		}

		if !path.IsAbs(input) {
			p, err := filepath.Abs(input)
			if err != nil {
				fmt.Println(Magenta("convert input path to abs failed"))
				os.Exit(1)
			}
			input = p
		}

		if !path.IsAbs(output) {
			output = path.Join(common.CurrentDir, path.Base(output))
		}

		coordinates = append(coordinates, args[2:]...)

		//fmt.Println("[debug] input:", input)
		//fmt.Println("[debug] output:", output)
		//fmt.Println("[debug] coordinate:", coordinates, ",", len(coordinates))
		//fmt.Println("[debug] concurrency:", concurrency)

		e := coordinate.NewExtractorCoordinate(input, output, coordinates, concurrency)
		if err := pdfextract.Extract(e); err != nil {
			fmt.Println(Magenta("Extract from pdf voucher failed,err:"), err)
		}
	},
}

func init() {
	pdfextractCmd.AddCommand(coordinateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// commonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// commonCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	coordinateCmd.Flags().IntP(common.CoordinateExtractFlagConcurrency, "c", 1, "分N组并发解析")
	viper.BindPFlag(common.CoordinateExtractFlagConcurrency, coordinateCmd.Flags().Lookup(common.CoordinateExtractFlagConcurrency))
}
