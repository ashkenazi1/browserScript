BrowserScript
=============

BrowserScript is a Go package that simplifies browser automation using [Chromedp](https://github.com/chromedp/chromedp). It allows you to execute browser automation scripts with actions like navigation, waiting for elements, evaluating JavaScript, clicking, and taking screenshots.

Features
--------

*   Automate browser interactions with Chromedp
*   Execute custom JavaScript
*   Navigate, click, set values, and take screenshots
*   Hide unwanted elements (popups, banners, etc.)
*   Headless browser support
*   Fully customizable tasks and actions

Installation
------------

To install the package, run:

    go get github.com/ashkenazi1/browserScript

Usage
-----

Here is an example of how to use BrowserScript to execute a simple script:

    
    package main
    
    import (
        "fmt"
        "time"
        "github.com/ashkenazi1/browserScript"
    )
    
    func main() {
        // Define a script with actions
        script := browserScript.Script{
            Name: "Example Script",
            Actions: []browserScript.Action{
                {Action: "navigate", Params: map[string]interface{}{"url": "https://example.com"}},
                {Action: "waitReady", Params: map[string]interface{}{"selector": "body"}},
                {Action: "getText", Params: map[string]interface{}{"selector": "h1", "result": "pageTitle"}},
            },
        }
    
        // Execute the script with a timeout and save screenshots
        err := browserScript.ExecuteScript(script, 30*time.Second, "./screenshots")
        if err != nil {
            fmt.Println("Error:", err)
        }
    }
    

Contributing
------------

Contributions, issues, and feature requests are welcome. Feel free to check the [issues page](https://github.com/ashkenazi1/browserScript/issues).

License
-------

Distributed under a free license. do whatever you want with it i don't care
