package reflectplus

import "fmt"

func PositionalError(p Positional, causedBy error) error {
	return fmt.Errorf("%s: %w", p.Position().ideString(), causedBy)
}
