package catalogue

type listRequest struct {
	Tags     []string `json:"tags"`
	Order    string   `json:"order"`
	PageNum  int      `json:"pageNum"`
	PageSize int      `json:"pageSize"`
}

type listResponse struct {
	Socks []Sock `json:"sock"`
	Err   error  `json:"err"`
}

type countRequest struct {
	Tags []string `json:"tags"`
}

type countResponse struct {
	N   int   `json:"size"`
	Err error `json:"err"`
}

type getRequest struct {
	ID string `json:"id"`
}

type getResponse struct {
	Sock Sock  `json:"sock"`
	Err  error `json:"Err"`
}

type tagsRequest struct{}
type tagsResponse struct {
	Tags []string `json:"tags"`
	Err  error    `json:"err"`
}

type healthRequest struct{}
type healthResponse struct {
	Health []Health `json:"health"`
}
