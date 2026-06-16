package handlers

type GetLinksDTO struct {
	Range string `form:"range"`
}

type CreateLinkDTO struct {
	OriginalUrl string  `json:"original_url" binding:"required,url"`
	ShortName   *string `json:"short_name"`
}

// TODO: check validation requirements (now both fields are optional)
type UpdateLinkDTO struct {
	OriginalUrl *string `json:"original_url" binding:"omitempty,url"`
	ShortName   *string `json:"short_name"`
}

type LinkResponse struct {
	ID          int64  `json:"id"`
	OriginalUrl string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortUrl    string `json:"short_url"`
}
