package browserScript

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type Config struct {
	Headless      bool
	Timeout       time.Duration
	ScreenshotDir string
	UserAgent     string
	Language      string
	RunMode       string
}

type ChromeVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type BrowserScript struct {
	cfg      Config
	Chromedp context.Context
}

func New(cfg Config) *BrowserScript {
	if cfg.UserAgent == "" {
		cfg.UserAgent = fmt.Sprintf("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", fetchLatestUserAgent())
	}
	if cfg.Language == "" {
		cfg.Language = "en-US"
	}
	if cfg.ScreenshotDir == "" {
		cfg.ScreenshotDir = "screenshots"
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", cfg.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("window-size", "1920,1080"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("lang", cfg.Language),
		chromedp.Flag("intl.accept_languages", cfg.Language),
		chromedp.Flag("accept-language", cfg.Language),
		chromedp.UserAgent(cfg.UserAgent),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)

	switch cfg.RunMode {
	case "interactive":
		// Create a persistent browser instance
		ctx, _ := chromedp.NewContext(allocCtx)
		return &BrowserScript{
			cfg:      cfg,
			Chromedp: ctx,
		}
	default:
		return &BrowserScript{
			cfg: cfg,
		}
	}
}

func fetchLatestUserAgent() string {
	response, err := http.Get("https://versionhistory.googleapis.com/v1/chrome/platforms/win/channels/stable/versions")
	if err != nil {
		return ""
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return ""
	}

	var data []ChromeVersion
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}

	latestVersion := data[5].Version
	return latestVersion
}

func (bs *BrowserScript) ExecuteScript(script Script) (results map[string]*string, err error) {

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("window-size", "1920,1080"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("lang", bs.cfg.Language),
		chromedp.Flag("intl.accept_languages", "en-US,en"),
		chromedp.Flag("accept-language", "en-US"),
		chromedp.UserAgent(bs.cfg.UserAgent),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	tasks := chromedp.Tasks{}
	results = make(map[string]*string)
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
		case "getLink":
			tempResult := new(string)
			results[action.Result] = tempResult
			tasks = append(tasks, chromedp.AttributeValue(action.Selector, "href", tempResult, nil))
		case "getImage":
			tempResult := new(string)
			results[action.Result] = tempResult
			tasks = append(tasks, chromedp.AttributeValue(action.Selector, "src", tempResult, nil))
		case "waitVisible":
			// Wait for an element to be visible
			tasks = append(tasks, chromedp.WaitVisible(action.Selector))
		case "waitReady":
			tasks = append(tasks, chromedp.WaitReady(action.Selector))
		case "getHtml":
			tempResult := new(string)
			results[action.Result] = tempResult
			tasks = append(tasks, chromedp.OuterHTML(action.Selector, tempResult))
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
				if err := os.MkdirAll(bs.cfg.ScreenshotDir, os.ModePerm); err != nil {
					return fmt.Errorf("failed to create screenshot directory: %v", err)
				}

				fileName := fmt.Sprintf("%s.%s", path, format)
				fullPath := filepath.Join(bs.cfg.ScreenshotDir, fileName)
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
			return results, fmt.Errorf("unknown action: %s", action.Action)
		}
	}

	// Run tasks
	if err := chromedp.Run(ctx, tasks); err != nil {
		return results, err
	}

	// // Process results
	// for key, value := range results {
	// 	fmt.Printf("%s: %s\n", key, *value)
	// }

	for key, value := range screenshotResults {
		fileName := filepath.Join(bs.cfg.ScreenshotDir, fmt.Sprintf("%s.png", key))
		if err := os.WriteFile(fileName, *value, 0644); err != nil {
			return results, fmt.Errorf("failed to write screenshot %s: %v", fileName, err)
		}
	}

	return results, nil
}

func (bs *BrowserScript) Close() {
	chromedp.Cancel(bs.Chromedp)
}
