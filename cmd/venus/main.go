package main

import (
	"fmt"

	"github.com/zan8in/gologger"
	"github.com/zan8in/venus/pkg/venus"
)

func main() {
	options := venus.ParseOptions()

	runner, err := venus.NewRunner(options)
	if err != nil {
		gologger.Fatal().Msg(err.Error())
	}

	options.OnResult = func(result map[string]string) {
		for key, value := range result {
			fmt.Println(key, value)
		}
	}

	if err := runner.Run(); err != nil {
		gologger.Fatal().Msg(err.Error())
	}

}
