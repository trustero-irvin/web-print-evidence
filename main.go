package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// Configurations
const (
	loginURL       = "https://login.microsoftonline.com"                                                        // Microsoft login page
	sharepointURL  = "https://trustero331.sharepoint.com/sites/Trustero-Sample-Tech-Wiki/SitePages/wdawda.aspx" // 🔹 Updated SharePoint URL
	cookieFile     = "cookies.json"
	screenshotFile = "screenshot.png"
	email          = "irvin@Trustero331.onmicrosoft.com" // 🔹 Updated Email
	password       = "Trustero2024!!"
	OTP            = "175694" // Replace with actual password

)

var shotCount = 1

func main() {
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // Set to true for headless mode
		chromedp.Flag("disable-gpu", true),
	)...)
	defer cancel()

	// Create a new browser context
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Increase timeout for operations
	ctx, cancel = context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	// Load and apply cookies if available
	// if err := loadCookies(ctx); err == nil {
	// 	fmt.Println("Loaded session cookies. Trying to access SharePoint...")
	// 	if isLoggedIn(ctx) {
	// 		fmt.Println("Successfully logged in using cookies!")
	// 		takeScreenshot(ctx) // Take screenshot after login success
	// 		return
	// 	}
	// 	fmt.Println("Cookies expired, performing full login...")
	// } else {
	// 	fmt.Println("No valid cookies found, performing full login...")
	// }

	// Perform login if cookies are invalid
	if err := loginToSharePoint(ctx); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	// Save fresh cookies after login
	// if err := saveCookies(ctx); err != nil {
	// 	log.Printf("⚠️ Failed to save cookies: %v", err)
	// } else {
	// 	fmt.Println("Session cookies saved successfully! 💾")
	// }

	// Take a screenshot after login
	takeScreenshot(ctx)
}

// // **Check if the user is logged in by verifying SharePoint access**
// func isLoggedIn(ctx context.Context) bool {
// 	var title string
// 	err := chromedp.Run(ctx,
// 		chromedp.Navigate(sharepointURL),
// 		chromedp.WaitReady("body"),
// 		chromedp.Title(&title),
// 	)
// 	if err != nil {
// 		fmt.Println("Not logged in. Need to authenticate.")
// 		return false
// 	}

// 	// Example: Check if the page title contains "Sign In" (adjust for SharePoint)
// 	if strings.Contains(strings.ToLower(title), "sign in") {
// 		return false
// 	}

// 	return true
// }

// **Login Function for SharePoint**
func loginToSharePoint(ctx context.Context) error {
	emailSelector := `#i0116`    // Email input field
	passwordSelector := `#i0118` // Password input field

	otherAccount := "#moreOptions"
	nextButtonSelector := `#idSIButton9`  // "Next" button
	staySignedInSelector := `#idBtn_Back` // "No" on Stay Signed In?
	otpTextBox := "#idTxtBx_SAOTCC_OTC"
	otpBtn := "#idSubmit_SAOTCC_Continue"

	return chromedp.Run(ctx,
		loadCookies(),
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(otherAccount),
		chromedp.Click(otherAccount),
		chromedp.Sleep(4*time.Second),

		chromedp.WaitVisible(emailSelector),
		chromedp.Click(emailSelector),
		chromedp.SendKeys(emailSelector, email),
		chromedp.Click(nextButtonSelector),
		chromedp.Sleep(2*time.Second),

		chromedp.WaitVisible(passwordSelector),
		chromedp.SendKeys(passwordSelector, password),
		chromedp.Click(nextButtonSelector),
		chromedp.Sleep(2*time.Second),

		//OTP
		chromedp.WaitVisible(otpTextBox),
		chromedp.Click(otpTextBox),
		chromedp.SendKeys(otpTextBox, OTP),
		chromedp.WaitVisible(otpBtn),
		chromedp.Click(otpBtn),
		chromedp.Sleep(5*time.Second),

		// Handle "Stay Signed In?" popup (if it appears)
		chromedp.WaitVisible(staySignedInSelector, chromedp.ByID),
		chromedp.Click(staySignedInSelector),
		chromedp.Sleep(3*time.Second),

		// Ensure SharePoint home page loads
		chromedp.WaitReady("body"),
		chromedp.Sleep(4*time.Second),
	)

}

// **Take Screenshot of the SharePoint Page**
func takeScreenshot(ctx context.Context) {
	account_id_btn := "#O365_MainLink_Me"
	content := "#spPageCanvasContent"
	internals := "vpc_WebPart.PageTitle.internal.253ea220-e282-42bc-84c6-4ae920adab64"

	err := chromedp.Run(ctx,
		chromedp.Navigate(sharepointURL),
		chromedp.WaitReady("body"),
		chromedp.WaitVisible(account_id_btn),
		chromedp.WaitReady(content),
		chromedp.Click(content),
		chromedp.WaitVisible(account_id_btn),
		chromedp.Click(internals),
		chromedp.KeyEvent(kb.End),
		chromedp.Sleep(10*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithScale(0.80).WithPrintBackground(true).Do(ctx)
			if err != nil {
				return err
			}
			return os.WriteFile("sample.pdf", buf, 0644)
		}),
	)

	if err != nil {
		log.Fatalf("Failed to take screenshot: %v", err)
	}

	fmt.Println("Screenshot saved successfully as:", screenshotFile)
	shotCount = shotCount + 1
}

func loadCookies() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		// Read the cookie file
		data, err := os.ReadFile(cookieFile)
		if err != nil {
			return err
		}

		// Decode JSON into cookies array
		var cookies []*network.Cookie
		if err := json.Unmarshal(data, &cookies); err != nil {
			return err
		}

		// Apply each cookie to Chrome session
		for _, cookie := range cookies {
			var expires *cdp.TimeSinceEpoch
			if cookie.Expires > 0 {
				exp := cdp.TimeSinceEpoch(time.Unix(int64(cookie.Expires), 0)) // ✅ Correct conversion
				expires = &exp
			}

			c := chromedp.FromContext(ctx)

			err := network.SetCookie(cookie.Name, cookie.Value).
				WithDomain(cookie.Domain).
				WithPath(cookie.Path).
				WithExpires(expires). // ✅ Properly sets expiration
				WithHTTPOnly(cookie.HTTPOnly).
				WithSecure(cookie.Secure).
				WithSameSite(network.CookieSameSite(cookie.SameSite)).
				Do(cdp.WithExecutor(ctx, c.Target))
			if err != nil {
				log.Printf("Failed to set cookie %s: %v", cookie.Name, err)
			}
		}

		if err != nil {
			return err
		}
		return nil
	})
}

// **Save Cookies to File**
func saveCookies(ctx context.Context) error {
	// Retrieve cookies from the current Chrome session
	cookies, err := network.GetCookies().Do(ctx)
	if err != nil {
		return err
	}

	// Convert to JSON format
	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return err
	}

	// Save to file
	return os.WriteFile(cookieFile, data, 0644)
}
