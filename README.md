GoBrowserBot
============

GoBrowserBot is a Go package that simplifies browser automation using [Chromedp](https://github.com/chromedp/chromedp). It allows you to execute browser automation scripts with actions like navigation, waiting for elements, evaluating JavaScript, clicking, and taking screenshots.

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

    go get github.com/yourusername/gobrowserbot

Usage
-----

Here is an example of how to use GoBrowserBot to execute a simple script:

    
    package main
    
    import (
        "fmt"
        "time"
        "github.com/yourusername/gobrowserbot"
    )
    
    func main() {
        // Define a script with actions
        script := virtualbrowser.Script{
            Name: "Example Script",
            Actions: []virtualbrowser.Action{
                {Action: "navigate", Params: map[string]interface{}{"url": "https://example.com"}},
                {Action: "waitReady", Params: map[string]interface{}{"selector": "body"}},
                {Action: "getText", Params: map[string]interface{}{"selector": "h1", "result": "pageTitle"}},
            },
        }
    
        // Execute the script with a timeout and save screenshots
        err := virtualbrowser.ExecuteScript(script, 30*time.Second, "./screenshots")
        if err != nil {
            fmt.Println("Error:", err)
        }
    }
    

Contributing
------------

Contributions, issues, and feature requests are welcome. Feel free to check the [issues page](https://github.com/yourusername/gobrowserbot/issues).

License
-------

Distributed under the MIT License. See [LICENSE](LICENSE) for more information.