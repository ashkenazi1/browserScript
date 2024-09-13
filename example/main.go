package main

import (
	"time"

	"github.com/ashkenazi1/browserScript"
)

func main() {
	script := browserScript.Script{
		Name: "Example",
		Actions: []browserScript.Action{
			{Action: "navigate", Params: map[string]interface{}{"url": "https://example.com"}},
			{Action: "takeElementScreenshot", Params: map[string]interface{}{
				"selector": "h1",
				"result":   "header_screenshot",
			}},
		},
	}

	err := browserScript.ExecuteScript(script, 5*time.Second, "./screenshots")
	if err != nil {
		panic(err)
	}
}
