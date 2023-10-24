package v2

type GroupType int

type GroupInterface interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
}

type Group struct {
	Type GroupType 
	TrackID int 
	Nodes []Node 
}

func NewSatGroup(groupType GroupType, trackID int) *Group {
	return &Group{
		Type: groupType,
		TrackID: trackID,
		Nodes: make([]Node, 0),
	}
}

func NewOtherGroup(groupType GroupType) *Group {
	return &Group{
		Type: groupType,
		Nodes: make([]Node, 0),
	}
}

func (group Group) Len() int {
	return len(group.Nodes)
}

func (group Group) Less(i, j int) bool {
	return group.Nodes[i].InTrackID < group.Nodes[j].InTrackID
}

func (group Group) Swap(i, j int) {
	group.Nodes[i], group.Nodes[j] = group.Nodes[j], group.Nodes[i]
}