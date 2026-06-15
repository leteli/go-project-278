package handlers

type GetLinksDTO struct {
	Limit  int64 `form:"limit,default=20" binding:"min=1,max=100"`
	Offset int64 `form:"offset,default=0" binding:"min=0"`
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
