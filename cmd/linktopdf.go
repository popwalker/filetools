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
	"invtools/pkg/linktopdf"
	"invtools/pkg/util"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var linktopdfCmdExample = fmt.Sprintf("%s\n%s\n%s\n%s\n",
	fmt.Sprintf("%s linktopdf -i=input.csv -o=output.zip", appName),
	fmt.Sprintf("%s linktopdf -i=input.csv -o=output.zip -c=2 -z=true -t=chromedp", appName),
	fmt.Sprintf("%s linktopdf input.csv output.zip", appName),
	fmt.Sprintf("%s linktopdf input.csv output.zip -c=2 -z=true -t=wkhtmltopdf", appName),
)

// linktopdfCmd represents the linktopdf command
var linktopdfCmd = &cobra.Command{
	Use:   "linktopdf",
	Short: "Print PDF from link(url).",
	Long: `
This command accept .csv/.xlsx/.txt file which contains links and print to pdf file.
Support concurrency print, compress to zip, choose print tools such as chromedp or wkhtmltopdf.
`,
	Example: linktopdfCmdExample,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			inputFile    = viper.GetString(common.LinkToPdfFlagInput)
			outputFile   = viper.GetString(common.LinkToPdfFlagOutput)
			concurrency  = viper.GetInt(common.LinkToPdfFlagConcurrency)
			needCompress = viper.GetBool(common.LinkToPdfFlagZip)
		)

		//fmt.Printf("-->>args:%v\n", args)
		//fmt.Printf("-->>flags:[%s],[%s],[%d]\n", inputFile, outputFile, concurrency)

		switch {
		case len(args) == 0: // full flag mode

		case len(args) == 2 && concurrency != 1: // args & flag mode
			inputFile = args[0]
			outputFile = args[1]
		case len(args) == 2 && concurrency == 1: // full args mode
			inputFile = args[0]
			outputFile = args[1]
		}

		// execute
		if err := linktopdf.Execute(inputFile, outputFile, concurrency, needCompress); err != nil {
			util.Printf("execute failed,err:%v", err)
		}

	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			var (
				inputFile  = viper.GetString(common.LinkToPdfFlagInput)
				outputFile = viper.GetString(common.LinkToPdfFlagOutput)
			)
			if inputFile == "" || outputFile == "" {
				fmt.Println(aurora.Magenta("Please Enter input file or output file"))
				//cmd.Help()
				os.Exit(1)
			}

			return nil
		}

		if len(args) != 2 {
			fmt.Println(aurora.Magenta("number of arguments invalid"))
			//cmd.Help()
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(linktopdfCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// linktopdfCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// linktopdfCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	linktopdfCmd.Flags().StringP(common.LinkToPdfFlagInput, "i", "", "input file, support .csv/.xlsx/.txt file")
	viper.BindPFlag(common.LinkToPdfFlagInput, linktopdfCmd.Flags().Lookup(common.LinkToPdfFlagInput))

	linktopdfCmd.Flags().StringP(common.LinkToPdfFlagOutput, "o", common.HomeDir, "output directory, (default current directory)")
	viper.BindPFlag(common.LinkToPdfFlagOutput, linktopdfCmd.Flags().Lookup(common.LinkToPdfFlagOutput))

	linktopdfCmd.Flags().IntP(common.LinkToPdfFlagConcurrency, "c", 1, "concurrency")
	viper.BindPFlag(common.LinkToPdfFlagConcurrency, linktopdfCmd.Flags().Lookup(common.LinkToPdfFlagConcurrency))

	linktopdfCmd.Flags().BoolP(common.LinkToPdfFlagZip, "z", false, "use zip to compress pdf, (default false)")
	viper.BindPFlag(common.LinkToPdfFlagZip, linktopdfCmd.Flags().Lookup(common.LinkToPdfFlagZip))

	linktopdfCmd.Flags().StringP(common.LinkToPdfFlagPrintType, "t", "chromedp", "use what kind of tool to print pdf, support chromedp and wkhtmltopdf")
	viper.BindPFlag(common.LinkToPdfFlagPrintType, linktopdfCmd.Flags().Lookup(common.LinkToPdfFlagPrintType))

	viper.Set(common.RunningDetective, linktopdf.LinktopdfName)
}
