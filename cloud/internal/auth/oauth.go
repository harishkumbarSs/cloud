package auth

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

var (
	googleOauthConfig *oauth2.Config
	oauthStateString  = "random-state"
	store            *sessions.CookieStore
)

func Init() {
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func GetGoogleUserInfo(client *http.Client) (*GoogleUserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state != "state" {
		log.Printf("Invalid oauth state: %s", state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := googleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("Code exchange failed: %s", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	client := googleOauthConfig.Client(r.Context(), token)
	userInfo, err := GetGoogleUserInfo(client)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	session, _ := store.Get(r, "session")
	session.Values["email"] = userInfo.Email
	session.Values["name"] = userInfo.Name
	session.Save(r, w)

	log.Printf("Successfully authenticated user: %s (%s)", userInfo.Name, userInfo.Email)
	http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
}
