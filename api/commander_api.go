package api

type Commander struct {
	ConstantPath string
}

func NewCommander(path string) *Commander {
	return &Commander{ ConstantPath: path }
}

func (c *Commander) ChangeCommanderPath(path string) {
	c.ConstantPath = path
}