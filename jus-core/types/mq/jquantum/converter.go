package jquantum

// Element 定义与JSON对应的数据结构
type Element interface{}

type PatternContent struct {
	Content []Element `json:"content"`
	Count   int64     `json:"count,omitempty"`
	Total   int64     `json:"total"`
}

type ResultJSON struct {
	Shots     int64                     `json:"shots"`
	NumQubits int64                     `json:"num_qubits"`
	Patterns  map[string]PatternContent `json:"patterns,omitempty"` // 添加omitempty
	Sequence  []Element                 `json:"sequence"`
}
