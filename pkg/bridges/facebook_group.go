package bridges

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/mmcdole/gofeed/atom"
)

const LOGIN_PAGE = "https://m.facebook.com/login/?fl"

type FacebookGroup struct {
	url          string
	entries      []atom.Entry
	last_fetched time.Time
	cache_time   time.Duration
}

func NewFacebookGroup(url string, cache_time time.Duration) FacebookGroup {
	return FacebookGroup{
		url:          url,
		entries:      []atom.Entry{},
		last_fetched: time.Now(),
		cache_time:   cache_time,
	}
}

func (group FacebookGroup) Entries() []atom.Entry {
	return group.entries
}

func (group FacebookGroup) LastFetched() time.Time {
	return group.last_fetched
}

func (group FacebookGroup) CacheTime() time.Duration {
	return group.cache_time
}

func (group FacebookGroup) UpdateEntries() error {
	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), "ws://localhost:9222/")
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	err := group.login(ctx)
	if err != nil {
		return err
	}

	err = group.loadGroup(ctx)
	return err
}

func (group FacebookGroup) login(ctx context.Context) error {
	var title string

	err := chromedp.Run(ctx, chromedp.Navigate(LOGIN_PAGE), chromedp.Title(&title))
	if err != nil {
		return err
	}

	if title == "Log into Facebook | Facebook" {
		println("Handling login")
		err = group.handle_login(ctx)
		if err != nil {
			return err
		}
	}

	err = chromedp.Run(ctx, chromedp.Title(&title))
	if err != nil {
		return err
	}

	if title == "Enter login code to continue" {
		println("Handling 2fa")
		err = group.handle_2fa(ctx)
		if err != nil {
			return err
		}
	}

	err = chromedp.Run(ctx, chromedp.Title(&title))
	if err != nil {
		return err
	}

	if title == "Remember browser" {
		println("Handling remember browser")
		err = group.handleRemember(ctx)
		if err != nil {
			return err
		}
	}

	println("Done login")
	return savePage(ctx, "done_login.html")
}

func (group FacebookGroup) handle_login(ctx context.Context) error {
	user := os.Getenv("KAYAK_FB_USER")
	pass := os.Getenv("KAYAK_FB_PASS")
	if user == "" || pass == "" {
		return errors.New("No Facebook credentials")
	}

	tasks := chromedp.Tasks{
		chromedp.WaitVisible("#m_login_email", chromedp.ByID),
		chromedp.SendKeys("#m_login_email", user, chromedp.ByID),
		chromedp.WaitVisible("#m_login_password", chromedp.ByID),
		chromedp.SendKeys("#m_login_password", pass, chromedp.ByID),
		chromedp.WaitVisible("#login_password_step_element", chromedp.ByID),
		chromedp.Click("#login_password_step_element", chromedp.ByID),
		chromedp.WaitNotPresent("#login_password_step_element", chromedp.ByID),
	}

	err := chromedp.Run(ctx, tasks)
	if err != nil {
		return err
	}

	return savePage(ctx, "login.html")
}

func (group FacebookGroup) handle_2fa(ctx context.Context) error {
	var auth_code string
	print("Enter 2FA code: ")
	fmt.Scanln(&auth_code)

	err := chromedp.Run(ctx,
		chromedp.WaitVisible("#approvals_code", chromedp.ByID),
		chromedp.SendKeys("#approvals_code", auth_code, chromedp.ByID),
		chromedp.WaitVisible("#checkpointSubmitButton-actual-button", chromedp.ByID),
		chromedp.Click("#checkpointSubmitButton-actual-button", chromedp.ByID))
	if err != nil {
		return err
	}

	return savePage(ctx, "2fa.html")
}

func (group FacebookGroup) handleRemember(ctx context.Context) error {
	err := chromedp.Run(ctx,
		chromedp.WaitVisible("#checkpointSubmitButton-actual-button", chromedp.ByID),
		chromedp.Click("#checkpointSubmitButton-actual-button", chromedp.ByID))
	if err != nil {
		return err
	}

	return savePage(ctx, "remember.html")
}

func (group FacebookGroup) loadGroup(ctx context.Context) error {
	println("Loading group")
	err := chromedp.Run(ctx, chromedp.Navigate(group.url))
	if err != nil {
		return err
	}

	return savePage(ctx, "group.html")
}

func savePage(ctx context.Context, filename string) error {
	var html string

	err := chromedp.Run(ctx,
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery))

	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, []byte(html), 0644)
}
