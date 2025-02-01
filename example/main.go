package main

import (
	"fmt"

	"github.com/ashkenazi1/browserScript"
)

func main() {

	cfg := browserScript.Config{
		RunMode:       "interactive",
		Language:      "en-US",
		UserAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36",
		ScreenshotDir: "./screenshots",
	}

	bs := browserScript.New(cfg)
	defer bs.Close()

	script := browserScript.Script{
		Name: "Simplified Example",
		Actions: []browserScript.Action{
			{Action: "navigate", Url: "https://termbin.com"},
			{Action: "waitVisible", Selector: "body"},
			{Action: "screenshot", Result: "full_screenshot3", Format: "png"},
			{Action: "getText", Selector: "h1", Result: "headerText"},
		},
	}

	results, err := bs.ExecuteScript(script)
	if err != nil {
		panic(err)
	}

	fmt.Println(*results["headerText"])
}
