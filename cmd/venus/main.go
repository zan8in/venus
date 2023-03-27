package main

import (
	"fmt"

	"github.com/zan8in/gologger"
	"github.com/zan8in/venus/pkg/result"
	"github.com/zan8in/venus/pkg/venus"
)

func main() {
	options := venus.ParseOptions()

	runner, err := venus.NewRunner(options)
	if err != nil {
		gologger.Fatal().Msg(err.Error())
	}

	runner.Run()

	options.OnResult = func(res result.Result) {
		for s := range res.GetDomainResults("lankecloud.com") {
			fmt.Println(s)
		}
	}
}
