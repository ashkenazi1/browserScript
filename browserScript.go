package browserScript

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Action struct {
	Action   string  `json:"action"`
	Url      string  `json:"url,omitempty"`
	Selector string  `json:"selector,omitempty"`
	Timeout  float64 `json:"timeout,omitempty"`
	Result   string  `json:"result,omitempty"`
	Path     string  `json:"path,omitempty"`
	Format   string  `json:"format,omitempty"`
	Value    string  `json:"value,omitempty"`
	Js       string  `json:"js,omitempty"`
}

type Script struct {
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}

func ExecuteScript(script Script, timeout time.Duration, screenshotDir string) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("window-size", "1920,1080"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	tasks := chromedp.Tasks{}
	results := make(map[string]*string)
	screenshotResults := make(map[string]*[]byte)

	// Handle JavaScript dialogs
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev.(type) {
		case *page.EventJavascriptDialogOpening:
			chromedp.Run(ctx, page.HandleJavaScriptDialog(true))
		}
	})

	for _, action := range script.Actions {
		switch action.Action {
		case "navigate":
			// Navigate to URL
			tasks = append(tasks, chromedp.Navigate(action.Url))
		case "waitVisible":
			// Wait for an element to be visible
			tasks = append(tasks, chromedp.WaitVisible(action.Selector))
		case "wait":
			// Wait for a specific time duration
			tasks = append(tasks, chromedp.Sleep(time.Duration(action.Timeout)*time.Second))
		case "screenshot":
			// Screenshot action
			path := action.Path
			if path == "" {
				path = action.Result
			}
			format := action.Format
			if format == "" {
				format = "png"
			}

			tempScreenshot := new([]byte)
			screenshotResults[action.Result] = tempScreenshot

			tasks = append(tasks, chromedp.FullScreenshot(tempScreenshot, 100))

			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				// Ensure the screenshot directory exists before saving the file
				if err := os.MkdirAll(screenshotDir, os.ModePerm); err != nil {
					return fmt.Errorf("failed to create screenshot directory: %v", err)
				}

				fileName := fmt.Sprintf("%s.%s", path, format)
				fullPath := filepath.Join(screenshotDir, fileName)
				return os.WriteFile(fullPath, *tempScreenshot, 0644)
			}))
		case "getText":
			// Get text from a specific element
			tempResult := new(string)
			results[action.Result] = tempResult
			tasks = append(tasks, chromedp.Text(action.Selector, tempResult))
		case "click":
			// Click on an element
			tasks = append(tasks, chromedp.Click(action.Selector))
		default:
			return fmt.Errorf("unknown action: %s", action.Action)
		}
	}

	// Run tasks
	if err := chromedp.Run(ctx, tasks); err != nil {
		return err
	}

	// Process results
	for key, value := range results {
		fmt.Printf("%s: %s\n", key, *value)
	}

	for key, value := range screenshotResults {
		fileName := filepath.Join(screenshotDir, fmt.Sprintf("%s.png", key))
		if err := os.WriteFile(fileName, *value, 0644); err != nil {
			return fmt.Errorf("failed to write screenshot %s: %v", fileName, err)
		}
	}

	return nil
}
