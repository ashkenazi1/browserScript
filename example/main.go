package main

import (
	"time"

	"github.com/ashkenazi1/browserScript"
)

func main() {
	// Extended script with various actions
	script := browserScript.Script{
		Name: "Extended Example",
		Actions: []browserScript.Action{
			// 1. Navigate to the website
			{Action: "navigate", Url: "https://example.com"},

			// 2. Wait until the body element is visible
			{Action: "waitVisible", Selector: "body"},

			// 3. Wait for 2 seconds
			{Action: "wait", Timeout: 2},

			// 4. Take a full-page screenshot and save it as full_screenshot.png
			{Action: "screenshot", Result: "full_screenshot", Path: "full_screenshot", Format: "png"},

			// 5. Capture a screenshot of the h1 element and save it as header_screenshot.png
			{Action: "takeElementScreenshot", Selector: "h1", Result: "header_screenshot", Path: "header_screenshot", Format: "png"},

			// 6. Get the text content of the h1 element and store it in the variable "headerText"
			{Action: "getText", Selector: "h1", Result: "headerText"},

			// 7. Click a button (assuming there is a button with the selector '#submitButton')
			{Action: "click", Selector: "#submitButton"},

			// 8. Set a value into a text input field (assuming there is an input with the selector '#username')
			{Action: "setValue", Selector: "#username", Value: "testuser"},

			// 9. Execute a custom JavaScript to log a message in the browser console
			{Action: "evaluate", Js: `console.log("Script executed!");`},
		},
	}

	// Execute the script with a timeout of 5 seconds and store screenshots in the ./screenshots directory
	err := browserScript.ExecuteScript(script, 5*time.Second, "./screenshots")
	if err != nil {
		panic(err)
	}
}
