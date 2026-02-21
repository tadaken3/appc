package api

// JSON:API response envelope
type Response[T any] struct {
	Data  []T            `json:"data"`
	Links PagingLinks    `json:"links,omitempty"`
	Meta  map[string]any `json:"meta,omitempty"`
}

type SingleResponse[T any] struct {
	Data T `json:"data"`
}

type PagingLinks struct {
	Self string `json:"self,omitempty"`
	Next string `json:"next,omitempty"`
}

type ResourceLinks struct {
	Self string `json:"self,omitempty"`
}
