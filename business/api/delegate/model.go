package delegate

import (
	"context"
	"fmt"
)

// Func represents a function that is registered and called by the system.
type Func func(context.Context, Data) error

// Data represents an event between domains.
type Data struct {
	Domain    string
	Action    string
	RawParams []byte
}

// String implements the Stringer interface.
func (d Data) String() string {
	return fmt.Sprintf(
		"Event{Domain:%#v, Action:%#v, RawParams:%#v}",
		d.Domain, d.Action, string(d.RawParams),
	)
}
