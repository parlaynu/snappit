package arena

type Archive interface {
	Baseline() string
	Manifest() string
	Data() string
}

type archive struct {
	baseline string
	manifest string
	data     string
}

func (a *archive) Baseline() string {
	return a.baseline
}

func (a *archive) Manifest() string {
	return a.manifest
}

func (a *archive) Data() string {
	return a.data
}
