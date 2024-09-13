package BrowserScript

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Action struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
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
			chromedp.Evaluate(`Object.defineProperty(navigator, 'languages', { get: () => ['en-US', 'en'] });`, nil).Do(ctx)
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
			url := action.Params["url"].(string)
			u := url
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				return chromedp.Navigate(u).Do(ctxWithTimeout)
			}))
		case "waitReady":
			selector := action.Params["selector"].(string)
			sel := selector
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				return chromedp.WaitReady(sel, chromedp.ByQuery).Do(ctxWithTimeout)
			}))
		case "waitForNavigation":
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				return chromedp.WaitReady("body", chromedp.ByQuery).Do(ctxWithTimeout)
			}))
		case "getText":
			selector, okSelector := action.Params["selector"].(string)
			resultVar, okResult := action.Params["result"].(string)

			if !okSelector || !okResult {
				return fmt.Errorf("invalid or missing 'selector' or 'result' in action.Params: %+v", action.Params)
			}

			tempResult := new(string)
			results[resultVar] = tempResult

			sel := selector
			resultPtr := tempResult

			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				// Check if element exists
				var nodes []*cdp.Node
				if err := chromedp.Nodes(sel, &nodes, chromedp.ByQuery).Do(ctxWithTimeout); err != nil {
					return err
				}
				if len(nodes) == 0 {
					return fmt.Errorf("no elements found for selector: %s", sel)
				}

				// Get the text content
				return chromedp.Text(sel, resultPtr, chromedp.NodeVisible).Do(ctxWithTimeout)
			}))
		case "click":
			selector := action.Params["selector"].(string)
			sel := selector
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				return chromedp.Click(sel, chromedp.ByQuery).Do(ctxWithTimeout)
			}))
		case "evaluate":
			js := action.Params["js"].(string)
			script := js
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				return chromedp.Evaluate(script, nil).Do(ctxWithTimeout)
			}))
		case "screenshot":
			resultVar, okResult := action.Params["result"].(string)
			if !okResult {
				return fmt.Errorf("invalid or missing 'result' in action.Params: %+v", action.Params)
			}
			tempScreenshot := new([]byte)
			screenshotResults[resultVar] = tempScreenshot

			screenshotPtr := tempScreenshot

			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				// Ensure the page is ready
				if err := chromedp.WaitReady("body", chromedp.ByQuery).Do(ctxWithTimeout); err != nil {
					return err
				}

				// Take the screenshot
				return chromedp.FullScreenshot(screenshotPtr, 70).Do(ctxWithTimeout)
			}))
		case "takeElementScreenshot":
			selector, okSelector := action.Params["selector"].(string)
			resultVar, okResult := action.Params["result"].(string)

			if !okSelector || !okResult {
				return fmt.Errorf("invalid or missing 'selector' or 'result' in action.Params: %+v", action.Params)
			}

			tempScreenshot := new([]byte)
			screenshotResults[resultVar] = tempScreenshot

			// Create local copies for closure
			sel := selector
			screenshotPtr := tempScreenshot

			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				// Check if element exists
				var nodes []*cdp.Node
				if err := chromedp.Nodes(sel, &nodes, chromedp.ByQuery).Do(ctxWithTimeout); err != nil {
					return err
				}
				if len(nodes) == 0 {
					return fmt.Errorf("no elements found for selector: %s", sel)
				}

				// Ensure the element is visible
				if err := chromedp.WaitVisible(sel, chromedp.ByQuery).Do(ctxWithTimeout); err != nil {
					return err
				}
				// Take the screenshot
				return chromedp.Screenshot(sel, screenshotPtr, chromedp.NodeVisible).Do(ctxWithTimeout)
			}))
		case "wait":
			duration := action.Params["timeout"].(float64)
			dur := time.Duration(duration) * time.Second
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				select {
				case <-time.After(dur):
					return nil
				case <-ctxWithTimeout.Done():
					return ctxWithTimeout.Err()
				}
			}))
		case "setValue":
			selector, okSelector := action.Params["selector"].(string)
			value, okValue := action.Params["value"].(string)
			if !okSelector || !okValue {
				return fmt.Errorf("invalid or missing 'selector' or 'value' in action.Params: %+v", action.Params)
			}
			sel := selector
			val := value
			tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				return chromedp.SetValue(sel, val, chromedp.ByQuery).Do(ctxWithTimeout)
			}))
		// ... [Other cases remain the same]
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
