package dtos

type MetaDataPagination struct {
	CurrentPage uint `json:"currentPage"`
	Limit       uint `json:"limit"`
	Offset      uint `json:"offset"`
	Total       uint `json:"total"`
	TotalPages  uint `json:"totalPages"`
	Previous    bool `json:"previous"`
	Next        bool `json:"next"`
}

type FilterPagination[I any] struct {
	Items    I                  `json:"items"`
	MetaData MetaDataPagination `json:"metaData"`
}
