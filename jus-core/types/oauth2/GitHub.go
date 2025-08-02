package oauth2

import "time"

type GitHubUserInfo struct {
	ID                int64   `json:"id"`
	Login             string  `json:"login"`
	Name              *string `json:"name"`
	AvatarURL         string  `json:"avatar_url"`
	Email             *string `json:"email"` // 指针处理 null
	NotificationEmail *string `json:"notification_email"`
}

type GitHubUserProfile struct {
	Login             string         `json:"login"`
	ID                int64          `json:"id"`
	NodeID            string         `json:"node_id"`
	AvatarURL         string         `json:"avatar_url"`
	GravatarID        string         `json:"gravatar_id"`
	URL               string         `json:"url"`
	HTMLURL           string         `json:"html_url"`
	FollowersURL      string         `json:"followers_url"`
	FollowingURL      string         `json:"following_url"`
	GistsURL          string         `json:"gists_url"`
	StarredURL        string         `json:"starred_url"`
	SubscriptionsURL  string         `json:"subscriptions_url"`
	OrganizationsURL  string         `json:"organizations_url"`
	ReposURL          string         `json:"repos_url"`
	EventsURL         string         `json:"events_url"`
	ReceivedEventsURL string         `json:"received_events_url"`
	Type              string         `json:"type"`
	UserViewType      string         `json:"user_view_type"`
	SiteAdmin         bool           `json:"site_admin"`
	Name              *string        `json:"name"` // 使用指针处理可能的 null
	Company           *string        `json:"company"`
	Blog              string         `json:"blog"`
	Location          *string        `json:"location"`
	Email             *string        `json:"email"`
	Hireable          *string        `json:"hireable"` // 根据 JSON 类型处理可能的 null
	Bio               string         `json:"bio"`
	TwitterUsername   *string        `json:"twitter_username"`
	NotificationEmail *string        `json:"notification_email"`
	PublicRepos       int            `json:"public_repos"`
	PublicGists       int            `json:"public_gists"`
	Followers         int            `json:"followers"`
	Following         int            `json:"following"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	PrivateGists      int            `json:"private_gists"`
	TotalPrivateRepos int            `json:"total_private_repos"`
	OwnedPrivateRepos int            `json:"owned_private_repos"`
	DiskUsage         int            `json:"disk_usage"`
	Collaborators     int            `json:"collaborators"`
	TwoFactorAuth     bool           `json:"two_factor_authentication"`
	Plan              GitHubUserPlan `json:"plan"`
}

type GitHubUserPlan struct {
	Name          string `json:"name"`
	Space         int    `json:"space"`
	Collaborators int    `json:"collaborators"`
	PrivateRepos  int    `json:"private_repos"`
}
