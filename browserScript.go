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
	Url      string  `json:"url,omitempty"`      // For navigate
	Selector string  `json:"selector,omitempty"` // For waitVisible, getText, click, etc.
	Timeout  float64 `json:"timeout,omitempty"`  // For wait
	Result   string  `json:"result,omitempty"`   // For getText and screenshot
	Path     string  `json:"path,omitempty"`     // For screenshot path
	Format   string  `json:"format,omitempty"`   // For screenshot format
	Value    string  `json:"value,omitempty"`    // For setValue
	Js       string  `json:"js,omitempty"`       // For evaluate
}

type Script struct {
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}

func getDefaultTasks() chromedp.Tasks {
	return chromedp.Tasks{
		// Task 1: Modify navigator properties and window.chrome
		chromedp.ActionFunc(func(ctx context.Context) error {
			chromedp.Evaluate(`delete navigator.__proto__.webdriver;`, nil).Do(ctx)
			chromedp.Evaluate(`window.chrome = { runtime: {} };`, nil).Do(ctx)
			chromedp.Evaluate(`Object.defineProperty(navigator, 'plugins', { get: () => [1, 2, 3, 4, 5] });`, nil).Do(ctx)
			chromedp.Evaluate(`Object.defineProperty(navigator, 'languages', { get: () => ['en-US', 'en'] });`, nil)
			return nil
		}),

		// Task 2: Hide unwanted elements using custom JavaScript
		chromedp.ActionFunc(func(ctx context.Context) error {
			js := `
				const selectors = [
					'div[role="dialog"]',
					'.modal', '.overlay', '.popup', '.lightbox',
					'#popup', '#overlay', '#modal', '#lightbox',
					'.popup-overlay', '.modal-overlay', '.overlay-container',
					'.dialog', '.cookie-banner', '.cookie-consent',
					'.alert', '.notification', '.ad-banner', '.promo-banner',
					'.subscribe-popup', '.newsletter-signup',
					'.consent-banner', '.full-screen-overlay'
				];

				selectors.forEach(selector => {
					document.querySelectorAll(selector).forEach(el => {
						el.style.display = 'none';
					});
				});
			`
			chromedp.Evaluate(js, nil).Do(ctx)
			return nil
		}),
	}
}

func ExecuteScript(script Script, timeout time.Duration, screenshotDir string) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("window-size", "1920,1080"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	tasks := getDefaultTasks()
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
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				return chromedp.Navigate(action.Url).Do(ctxWithTimeout)
			}))
		case "waitVisible":
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				return chromedp.WaitVisible(action.Selector, chromedp.ByQuery).Do(ctxWithTimeout)
			}))
		case "wait":
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				time.Sleep(time.Duration(action.Timeout) * time.Second)
				return nil
			}))
		case "screenshot":
			// Get the path and format if they are provided
			path := action.Path
			if path == "" {
				path = action.Result // Default to using result as the file name
			}
			format := action.Format
			if format == "" {
				format = "png" // Default to PNG if no format is provided
			}

			tempScreenshot := new([]byte)
			screenshotResults[action.Result] = tempScreenshot

			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				if err := chromedp.WaitReady("body", chromedp.ByQuery).Do(ctxWithTimeout); err != nil {
					return err
				}
				return chromedp.FullScreenshot(tempScreenshot, 70).Do(ctxWithTimeout)
			}))

			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				fileName := fmt.Sprintf("%s.%s", path, format)
				fullPath := filepath.Join(screenshotDir, fileName)
				return os.WriteFile(fullPath, *screenshotResults[action.Result], 0644)
			}))
		case "getText":
			selector := action.Selector
			resultVar := action.Result

			tempResult := new(string)
			results[resultVar] = tempResult

			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				return chromedp.Text(selector, tempResult, chromedp.NodeVisible).Do(ctx)
			}))
		case "click":
			selector := action.Selector
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				return chromedp.Click(selector, chromedp.ByQuery).Do(ctx)
			}))
		case "evaluate":
			js := action.Js
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				return chromedp.Evaluate(js, nil).Do(ctx)
			}))
		case "setValue":
			selector := action.Selector
			value := action.Value
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				return chromedp.SetValue(selector, value, chromedp.ByQuery).Do(ctx)
			}))
		default:
			return fmt.Errorf("unknown action: %s", action.Action)
		}
	}

	// Run the tasks
	if err := chromedp.Run(ctx, tasks); err != nil {
		return err
	}

	// Ensure the screenshot directory exists
	if err := os.MkdirAll(screenshotDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create screenshot directory: %v", err)
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
