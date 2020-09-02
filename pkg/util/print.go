package util

import (
	"fmt"

	"invtools/common"

	"github.com/spf13/viper"
)

func Printf(format string, args ...interface{}) {
	fmt.Printf("[%s] %s\n", viper.Get(common.RunningDetective), fmt.Sprintf(format, args...))
	//output := fmt.Sprintf("[%s] %s", viper.Get(RunningDective), fmt.Sprintf(format, args...))
	//f := bufio.NewWriter(os.Stdout)
	//defer f.Flush()
	//f.WriteString(output)
}
