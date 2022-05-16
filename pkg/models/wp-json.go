package models

type WPJsonUsersResponse []struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Slug        string `json:"slug"`
	AvatarUrls  struct {
		Num24 string `json:"24"`
		Num48 string `json:"48"`
		Num96 string `json:"96"`
	} `json:"avatar_urls"`
	Meta  []interface{} `json:"meta"`
	Links struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
		Collection []struct {
			Href string `json:"href"`
		} `json:"collection"`
	} `json:"_links"`
}
