package types

type Label string

const (
	LabelApp   Label = "x-platform-app"
	LabelOwner Label = "x-platform-owner"
	LabelType  Label = "x-platform-type"
)

func (l Label) String() string {
	return string(l)
}
