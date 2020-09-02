package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"time"

	"invtools/utils/errors"
)

func ExtractTextByCoordinate(input, coordinate, output string) (string, error) {
	cmdTpl := fmt.Sprintf(`/usr/local/bin/tet -o %s --pageopt "includebox={{%s}}" %s`,
		output, coordinate, input)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdTpl)
	err := cmd.Run()
	if err != nil {
		return "", errors.Errorf(err, "tet extract text by coordinate failed, cmd:%s", cmdTpl)
	}

	//f, err := os.Open(output)
	//if err != nil {
	//    return "", errors.Errorf(err, "open tet output file failed, file:%s", output)
	//}
	//defer f.Close()

	b, err := ioutil.ReadFile(output)
	if err != nil {
		return "", errors.Errorf(err, "read tet output file failed, file:%s", output)
	}

	ret := StringPurify(string(b))
	return ret, nil
}
