package v2

type GroupType int

type GroupInterface interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
}

type Group struct {
	Type GroupType `json:"type,omitempty"`
	Nodes []Node `json:"nodes,omitempty"`
}

func (group Group) Len() int {
	return len(group.Nodes)
}

func (group Group) Less(i, j int) bool {
	return group.Nodes[i].Angle() < group.Nodes[j].Angle()
}

func (group Group) Swap(i, j int) {
	group.Nodes[i], group.Nodes[j] = group.Nodes[j], group.Nodes[i]
}