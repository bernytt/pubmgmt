package pub

type SubversionInfo struct {
	ID       uint64   `json:"id"`
	Name     string   `json:"name"`
	Dest     string   `json:"dest"`
	Repo     string   `json:"repo"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Revision string   `json:"revision"`
	Hosts    []string `json:"hosts"`
}

func (*SubversionInfo) UniqueFields() []string {
	return []string{"ID"}
}
