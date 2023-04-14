package main

import (
	"fmt"
	"sync"

	"github.com/zan8in/gologger"
	fileutil "github.com/zan8in/pins/file"
	timeutil "github.com/zan8in/pins/time"
	"github.com/zan8in/venus/pkg/venus"
)

func main() {

	options := venus.ParseOptions()

	runner, err := venus.NewRunner(options)
	if err != nil {
		gologger.Fatal().Msg(err.Error())
	}

	var sf fileutil.SafeFile
	wg := &sync.WaitGroup{}
	if len(options.Output) == 0 {
		options.Output = timeutil.Format(timeutil.Format_1) + ".txt"
	}

	wg.Add(1)
	options.OnResult = func(result map[string]string) {
		for key, value := range result {
			if key == "DONE" {
				wg.Done()
				return
			}
			fmt.Println(key, value)

			sf.Write(options.Output, value+"\r\n")
		}
	}

	if err := runner.Run(); err != nil {
		gologger.Fatal().Msg(err.Error())
	}

	wg.Wait()
}
