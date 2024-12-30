package types

type Label string

const (
	LabelApp    Label = "x-platform-app"
	LabelOwners Label = "x-platform-owners"
	LabelType   Label = "x-platform-type"
)

func (l Label) String() string {
	return string(l)
}
