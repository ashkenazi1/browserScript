package main

import (
	"time"

	"github.com/ashkenazi1/browserScript"
)

func main() {
	script := browserScript.Script{
		Name: "Simplified Example",
		Actions: []browserScript.Action{
			{Action: "navigate", Url: "https://termbin.com"},
			{Action: "waitVisible", Selector: "body"},
			{Action: "screenshot", Result: "full_screenshot3", Format: "png"},
			{Action: "getText", Selector: "h1", Result: "headerText"},
		},
	}

	err := browserScript.ExecuteScript(script, 5*time.Second, "./screenshots")
	if err != nil {
		panic(err)
	}
}
