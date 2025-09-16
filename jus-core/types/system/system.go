package system

const (
	jquantum = 1 << iota
	hermes
	//apollo
	//zeus
)

// SubsystemID JUS Sub System ID
var SubsystemID = struct {
	JQuantum int
	Hermes   int
}{
	JQuantum: jquantum,
	Hermes:   hermes,
}

var SubsystemScopes = map[int]struct{}{
	jquantum: {},
}
