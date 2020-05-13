package reflectplus

import "fmt"

func PositionalError(p Positional, causedBy error) error {
	if causedBy != nil {
		return fmt.Errorf("%s: %w", p.Position().ideString(), causedBy)
	}
	return fmt.Errorf("%s", p.Position().ideString())
}
