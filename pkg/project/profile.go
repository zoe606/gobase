package project

import "fmt"

type Profile string

const (
	Full    Profile = "full"
	Minimal Profile = "minimal"
)

func (p Profile) Validate() error {
	switch p {
	case Full, Minimal:
		return nil
	default:
		return fmt.Errorf("unsupported project profile %q; choose full or minimal", p)
	}
}
