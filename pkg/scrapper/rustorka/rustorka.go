package rustorka

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/vanyason/infozone-gommaraizer/pkg/logger"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	ForumURL = "https://rustorka.com/forum/"
	LoginURL = "https://rustorka.com/forum/login.php"

	UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:129.0) Gecko/20100101 Firefox/129.0"
)

// GetRustorkaAuthToken returns a cookie jar with the Rustorka auth token or an error.
// It sends a POST request to the login page with the username and password as a
// URL-encoded form, and then follows redirects (if any) to get the final response.
// The response is expected to contain the auth token in the Set-Cookie header,
// which is stored in the cookie jar and returned.
func GetRustorkaAuthToken() (*cookiejar.Jar, error) {
	logger.Info("Getting Rustorka auth token")

	// Create a new HTTP request
	const authPayload = "login_username=asdasd&login_password=asdasd&autologin=1&login=%C2%F5%EE%E4"
	req, err := http.NewRequest("POST", LoginURL, strings.NewReader(authPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Create a cookie jar to store cookies across redirects
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// Setup a client with the cookie jar and redirect handler
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 { //< Merge cookies from previous responses into the new request
				previousResponse := via[len(via)-1]
				mergedCookies := jar.Cookies(previousResponse.URL)
				for _, cookie := range mergedCookies {
					req.AddCookie(cookie)
				}
			}
			return nil //< Follow the redirect
		},
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	return jar, nil
}

// FetchHTML fetches HTML from the given URL, using the provided cookie jar for auth.
// The function returns the HTML as a string, or an error if something goes wrong.
// The cookie jar is required, since it holds authentication data.
func FetchHTML(url string, jar *cookiejar.Jar) (string, error) {
	logger.Info("Fetching HTML from", "url", url)

	// Create an HTTP GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", UserAgent)

	// Add cookie jar
	client := &http.Client{Jar: jar}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	// Convert the response body from windows-1251 to UTF-8 if needed
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Type") == "text/html; charset=windows-1251" {
		logger.Debug("Converting response body from windows-1251 to UTF-8")
		reader = transform.NewReader(resp.Body, charmap.Windows1251.NewDecoder())
	}

	// Read the response body
	body, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("error reading the response body: %w", err)
	}

	// Return the HTML as a string
	return string(body), nil
}
