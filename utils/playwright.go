package utils

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/math"

	"github.com/PuerkitoBio/goquery"
	"github.com/mattn/go-runewidth"
	"github.com/playwright-community/playwright-go"
	log "github.com/sirupsen/logrus"
)

// ScreenShotPageOption 截屏选项
type ScreenShotPageOption struct {
	Width    int
	Height   int
	DPI      float64
	Before   func(page playwright.Page)
	PwOption playwright.PageScreenshotOptions
	Sleep    time.Duration
}

// ScreenShotElementOption 元素截屏选项
type ScreenShotElementOption struct {
	Width    int
	Height   int
	DPI      float64
	Before   func(page playwright.Page)
	PwOption playwright.LocatorScreenshotOptions
	Sleep    time.Duration
}

var (
	// GlobalCSS 全局样式, 用于屏蔽浮动/广告元素等
	GlobalCSS = `
#ageDisclaimerMainBG,
.desktop-dialog-open,
#modalWrapMTubes {
    display: none !important;
    visibility: hidden !important;
}
`
	pw *playwright.Playwright
	// ctx    playwright.BrowserContext
	inited = false
	// DefaultPageOptions 默认截图选项
	DefaultPageOptions = playwright.PageScreenshotOptions{
		FullPage:   playwright.Bool(true),
		Type:       playwright.ScreenshotTypeJpeg,
		Quality:    playwright.Int(70),
		Timeout:    playwright.Float(10_000),
		Animations: playwright.ScreenshotAnimationsAllow,
		Scale:      playwright.ScreenshotScaleDevice,
		Style:      playwright.String(GlobalCSS + "\n" + ``),
	}
	// DefaultElementOptions 默认元素截屏选项
	DefaultElementOptions = playwright.LocatorScreenshotOptions{
		Type:       playwright.ScreenshotTypeJpeg,
		Quality:    playwright.Int(70),
		Timeout:    playwright.Float(10_000),
		Animations: playwright.ScreenshotAnimationsAllow,
		Scale:      playwright.ScreenshotScaleDevice,
		Style:      playwright.String(GlobalCSS + "\n" + `body{padding: 0;margin: 0;}`),
	}
)

func init() {
	t := time.NewTicker(3 * time.Second)
	go func() {
		for range t.C {
			var err error
			log.Infoln("[playwright] 启动中...")
			pw, err = playwright.Run(&playwright.RunOptions{
				SkipInstallBrowsers: false,
				Browsers:            []string{"chromium-headless-shell"},
			})
			log.Infoln("[playwright] 启动成功！！！")
			if err != nil {
				pw, err = playwright.Run(&playwright.RunOptions{
					SkipInstallBrowsers: false,
					Browsers:            []string{"chromium-headless-shell"},
				})
				log.Warnln("[playwright] 启动失败，正在初始化！！！")
				err = playwright.Install(&playwright.RunOptions{
					SkipInstallBrowsers: false,
					Browsers:            []string{"chromium-headless-shell"},
				})
				if err != nil {
					log.Errorln("[playwright] 无法初始化！！！", err)
					continue
				}
			}
			inited = true
			return
		}
	}()
}

// WaitImage 等待图片加载完毕
func WaitImage(page playwright.Page) {
	all, _ := page.Locator("img").All()
	for _, locator := range all {
		if visible, err := locator.IsVisible(); err != nil || !visible {
			continue
		}
		_ = locator.ScrollIntoViewIfNeeded(playwright.LocatorScrollIntoViewIfNeededOptions{Timeout: playwright.Float(500)})
		_, _ = locator.Evaluate(`(e) => {
    if (e.offsetParent) {
        e.offsetParent.scrollLeft = 0;
        e.offsetParent.scrollTop = 0;
    }
}`, playwright.LocatorEvaluateOptions{Timeout: playwright.Float(500)})
		_ = playwright.NewPlaywrightAssertions(2000).Locator(locator).ToHaveJSProperty("complete", true)
	}
	_, _ = page.Evaluate(`document.querySelectorAll("*").forEach(e=>{e.scrollLeft = 0;e.scrollTop = 0;})`)
}

// Clean 清理页面，移除浮动元素等
func Clean(page playwright.Page) {
	file, err := os.ReadFile("clean.js")
	if err != nil {
		return
	}
	_, _ = page.Evaluate(string(file))
}

var debug = false

func launchContext(dpi float64) (playwright.BrowserContext, error) {
	if dpi == 0 {
		dpi = 1.5
	}
	var err error
	debug, err = strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		debug = false
	}
	log.Debugf("[playwright] 正在创建chromium, dpi:%.1f debug:%v", dpi, debug)
	context, err := pw.Chromium.LaunchPersistentContext(fmt.Sprintf("./bw%.1f", dpi), playwright.BrowserTypeLaunchPersistentContextOptions{
		DeviceScaleFactor: playwright.Float(dpi),
		ChromiumSandbox:   playwright.Bool(false),
		AcceptDownloads:   playwright.Bool(false),
		Headless:          playwright.Bool(true),
	})
	if err == nil {
		_ = context.Route("https://*.qlogo.cn/**", func(route playwright.Route) {
			_ = route.Continue()
		})
	} else {
		log.Warnf("[playwright] 无法创建chromium, err:%v", err)
	}
	return context, err
}
func getScreenShotPageOption(option []ScreenShotPageOption) ScreenShotPageOption {
	o := ScreenShotPageOption{Width: 600, Sleep: time.Millisecond * 100, PwOption: DefaultPageOptions}
	if len(option) != 0 {
		o = option[0]
	}
	return o
}

func getScreenShotElementOption(option []ScreenShotElementOption) ScreenShotElementOption {
	o := ScreenShotElementOption{Width: 600, Sleep: time.Millisecond * 100, PwOption: DefaultElementOptions}
	if len(option) != 0 {
		o = option[0]
	}
	return o
}

// ScreenShotPageContent takes a string of HTML content and optionally a set of screen shot options to capture a page screenshot.
// This function initializes the browser context, creates a new page, sets the page content, and then takes a screenshot according to the specified options.
func ScreenShotPageContent(content string, option ...ScreenShotPageOption) (bytes []byte, err error) {
	// Check if Playwright has been initialized
	if !inited {
		log.Error("Playwright not initialized")
		return nil, errors.New("playwright not inited")
	}

	// Get the screen shot options, using the default if none are provided
	o := getScreenShotPageOption(option)

	// Launch a new browser context
	ctx, err := launchContext(o.DPI)
	if err != nil {
		log.Errorf("Failed to launch context: %v", err)
		return nil, err
	}
	if !debug {
		defer ctx.Close()
	}

	// Create a new page in the browser context
	page, err := ctx.NewPage()
	if err != nil {
		log.Errorf("Failed to create page: %v", err)
		return nil, err
	}
	if !debug {
		defer page.Close()
	}

	// Set the content of the page
	err = page.SetContent(content)
	if err != nil {
		log.Errorf("Failed to set page content: %v", err)
		return nil, err
	}

	// Take a screenshot of the page with the specified options
	return screenShotPage(page, o.Width, o.Height, o.Sleep, o.Before, o.PwOption)
}

// ScreenShotElementContent takes a string of HTML content, a CSS selector for the target element, and optionally a set of screen shot options to capture a screenshot of a specific element.
// This function initializes the browser context, creates a new page, sets the page content, and then takes a screenshot of the specified element according to the options.
func ScreenShotElementContent(content string, selector string, option ...ScreenShotElementOption) (bytes []byte, err error) {
	// Check if Playwright has been initialized
	if !inited {
		log.Error("Playwright not initialized")
		return nil, errors.New("playwright not inited")
	}

	// Get the screen shot options for the element, using the default if none are provided
	o := getScreenShotElementOption(option)

	// Launch a new browser context
	ctx, err := launchContext(o.DPI)
	if err != nil {
		log.Errorf("Failed to launch context: %v", err)
		return nil, err
	}
	if !debug {
		defer ctx.Close()
	}

	// Create a new page in the browser context
	page, err := ctx.NewPage()
	if err != nil {
		log.Errorf("Failed to create page: %v", err)
		return nil, err
	}
	if !debug {
		defer page.Close()
	}

	// Set the content of the page
	err = page.SetContent(content)
	if err != nil {
		log.Errorf("Failed to set page content: %v", err)
		return nil, err
	}

	// Log the start of taking a screenshot
	log.Info("Taking screenshot")

	// Take a screenshot of the specified element with the specified options
	return screenShotElement(page, selector, o.Width, o.Height, o.Sleep, o.Before, o.PwOption)
}

// ScreenShotPageURL 网址截屏
func ScreenShotPageURL(u string, option ...ScreenShotPageOption) (bytes []byte, err error) {
	if !inited {
		return nil, errors.New("playwright not inited")
	}
	parse, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	ipReg := regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)
	if parse.Scheme == "" {
		if ipReg.MatchString(parse.Host) {
			parse.Scheme = "http"
		} else {
			parse.Scheme = "https"
		}
	}
	if parse.Scheme != "https" && parse.Scheme != "http" {
		return nil, errors.New("unsupport schema")
	}
	o := getScreenShotPageOption(option)
	ctx, err := launchContext(o.DPI)
	if err != nil {
		log.Errorf("Failed to launch context: %v", err)
		return nil, err
	}
	if !debug {
		defer ctx.Close()
	}
	page, err := ctx.NewPage()
	if err != nil {
		return nil, err
	}
	if !debug {
		defer page.Close()
	}
	response, err := page.Goto(parse.String(), playwright.PageGotoOptions{Timeout: DefaultPageOptions.Timeout})
	if errors.Is(err, playwright.ErrTimeout) {
		log.Warnln(err)
		response, err = page.Goto(parse.String())
	}
	if err != nil {
		log.Warnln(err)
		return nil, err
	}
	if !response.Ok() {
		if debug {
			log.Warningf("response %d", response.StatusText())
		} else {
			return nil, errors.New("response not ok")
		}
	}
	return screenShotPage(page, o.Width, o.Height, o.Sleep, o.Before, o.PwOption)
}

// ScreenShotElementURL 网址元素截屏
func ScreenShotElementURL(u string, selector string, option ...ScreenShotElementOption) (bytes []byte, err error) {
	if !inited {
		return nil, errors.New("playwright not inited")
	}
	parse, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	ipReg := regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)
	if parse.Scheme == "" {
		if ipReg.MatchString(parse.Host) {
			parse.Scheme = "http"
		} else {
			parse.Scheme = "https"
		}
	}
	if parse.Scheme != "https" && parse.Scheme != "http" {
		return nil, errors.New("unsupport schema")
	}
	o := getScreenShotElementOption(option)
	ctx, err := launchContext(o.DPI)
	if err != nil {
		log.Errorf("Failed to launch context: %v", err)
		return nil, err
	}
	if !debug {
		defer ctx.Close()
	}
	page, err := ctx.NewPage()
	if err != nil {
		return nil, err
	}
	if !debug {
		defer page.Close()
	}
	response, err := page.Goto(parse.String())
	if errors.Is(err, playwright.ErrTimeout) {
		response, err = page.Goto(parse.String())
	}
	if err != nil {
		return nil, err
	}
	if debug {
		log.Warningf("response %d", response.StatusText())
	} else {
		return nil, errors.New("response not ok")
	}
	return screenShotElement(page, selector, o.Width, o.Height, o.Sleep, o.Before, o.PwOption)
}

// Truncate 截断字符串(按照打印出来的长度计算，相对于ascii字符宽度)
func Truncate(title string, maxlen int) string {
	return runewidth.Truncate(title, maxlen, "...")
}

var funcs = template.FuncMap{
	"truncate": Truncate,
	"escape": func(in string) template.HTML {
		return template.HTML(in)
	},
	"replace": func(reg string, repl string, src string) string {
		regex, err := regexp.Compile(reg)
		if err != nil {
			log.Errorf("regexp compile error: %v", err)
			return src // 返回原始字符串
		}
		return regex.ReplaceAllString(src, repl)
	},
	"tohtml": func(in string) *goquery.Document {
		reader, err := goquery.NewDocumentFromReader(strings.NewReader(in))
		if err != nil {
			log.Errorf("tohtml error: %v", err)
			return nil // 返回空指针
		}
		return reader
	},
	"select": func(selector string, in string) *goquery.Selection {
		reader, err := goquery.NewDocumentFromReader(strings.NewReader(in))
		if err != nil {
			log.Errorf("select error: %v", err)
			return nil // 返回空指针
		}
		find := reader.Find(selector)
		return find
	},
	"selContent": func(in *goquery.Selection) template.HTML {
		return template.HTML(strings.Trim(in.Text(), " \n\r"))
	},
	"docContent": func(in *goquery.Document) template.HTML {
		return template.HTML(strings.Trim(in.Text(), " \n\r"))
	},
	"startWith": strings.HasPrefix,
	"endWith":   strings.HasSuffix,
	"isnil": func(obj interface{}) bool {
		return obj == nil
	},
	"notnil": func(obj interface{}) bool {
		return obj != nil
	},
}

// ScreenShotPageTemplate 模板截屏
func ScreenShotPageTemplate(name string, data any, option ...ScreenShotPageOption) (bytes []byte, err error) {
	if !inited {
		return nil, errors.New("playwright not inited")
	}
	t, err := template.New(name).Funcs(funcs).ParseGlob("template/**/*.*html")
	if err != nil {
		return nil, err
	}
	buf := strings.Builder{}
	err = t.Execute(&buf, data)
	if err != nil {
		return nil, err
	}
	return ScreenShotPageContent(buf.String(), option...)
}

// ScreenShotElementTemplate 元素模板截屏
func ScreenShotElementTemplate(name string, selector string, data any, option ...ScreenShotElementOption) (bytes []byte, err error) {
	if !inited {
		return nil, errors.New("playwright not inited")
	}
	t, err := template.New(name).Funcs(funcs).ParseGlob("template/**/*.*html")
	if err != nil {
		return nil, err
	}
	buf := strings.Builder{}
	err = t.Execute(&buf, data)
	if err != nil {
		return nil, err
	}
	return ScreenShotElementContent(buf.String(), selector, option...)
}

func screenShotPage(page playwright.Page, width, height int, sleep time.Duration, before func(page playwright.Page), pwOption playwright.PageScreenshotOptions) ([]byte, error) {
	if sleep > 0 {
		log.Infof("[playwright] 等待 %v ms", sleep)
		time.Sleep(sleep)
	}
	log.Infoln("[playwright] 清理页面元素")
	Clean(page)
	log.Infoln("[playwright] 等待图片加载")
	WaitImage(page)
	log.Infoln("[playwright] 准备执行用户自定义函数")
	if before != nil {
		before(page)
		log.Infoln("[playwright] 执行用户自定义函数完毕")
	}
	if height < 20 {
		height = 20
	}
	var start = time.Now()
	if *pwOption.FullPage {
		var count = 1
	redo:
		err := page.SetViewportSize(width, height)
		if err != nil {
			log.Errorf("Error setting viewport size: %v", err)
			return nil, err
		}
		time.Sleep(200 * time.Millisecond)

		evaluated, err := page.Evaluate(`document.documentElement.scrollHeight`)
		if err != nil {
			log.Errorf("Error evaluating document scroll height: %v", err)
			return nil, err
		}

		if math.Abs(height-evaluated.(int)) > 10 && count > 0 {
			height = evaluated.(int)
			count--
			goto redo
		}
		height = evaluated.(int)
	}
	var maxHeight = 5_0000
	if height > maxHeight {
		log.Infof("[playwright] 视窗高度(%d)超出最大限制(%d)", height, maxHeight)
		height = maxHeight
	}
	log.Infof("[playwright] 重设视窗大小%dx%d 耗时 %vms", width, height, time.Now().Sub(start).Milliseconds())
	err := page.SetViewportSize(width, height)
	if err != nil {
		log.Infof("Error setting viewport size after evaluation: %v", err)
		return nil, err
	}

	log.Infoln("[playwright] 准备截图页面")
	screenshot, err := page.Screenshot(pwOption)
	if err != nil {
		log.Infof("Error taking screenshot: %v", err)
		return nil, err
	}
	if len(screenshot) == 0 {
		return nil, errors.New("empty screenshot")
	}
	log.Infof("[playwright] 全屏截图结果%v,err:%v", len(screenshot), err)

	return screenshot, nil
}

func screenShotElement(page playwright.Page, selector string, width, height int, sleep time.Duration, before func(page playwright.Page), pwOption playwright.LocatorScreenshotOptions) ([]byte, error) {
	logger := log.New()

	Clean(page)
	WaitImage(page)
	if before != nil {
		before(page)
	}
	if height == 0 {
		height = 100
	}

	err := page.SetViewportSize(width, height)
	if err != nil {
		logger.WithError(err).Error("Error setting viewport size")
		return nil, err
	}

	evaluated, err := page.Evaluate(`document.documentElement.scrollHeight`)
	if err != nil {
		logger.WithError(err).Error("Error evaluating document scroll height")
		return nil, err
	}

	height = evaluated.(int)

	err = page.SetViewportSize(width, height)
	if err != nil {
		logger.WithError(err).Error("Error setting viewport size after evaluation")
		return nil, err
	}

	WaitImage(page)

	time.Sleep(sleep)

	locator := page.Locator(selector)
	screenshot, err := locator.Screenshot(pwOption)
	if err != nil {
		logger.WithError(err).Error("Error taking screenshot")
		return nil, err
	}

	return screenshot, nil
}
