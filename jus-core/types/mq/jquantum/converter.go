package jquantum

// 定义与JSON对应的数据结构
type Element interface{}

type PatternContent struct {
	Content []Element `json:"content"`
	Count   int       `json:"count,omitempty"`
	Total   int       `json:"total"`
}

type ResultJSON struct {
	NumQubits int                       `json:"num_qubits"`
	Patterns  map[string]PatternContent `json:"patterns,omitempty"` // 添加omitempty
	Sequence  []Element                 `json:"sequence"`
}
