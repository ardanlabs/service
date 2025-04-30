package userbus

import (
	"encoding/json"
	"fmt"

	"github.com/ardanlabs/service/business/sdk/delegate"
	"github.com/google/uuid"
)

// DomainName represents the name of this domain.
const DomainName = "user"

// Set of delegate actions.
const (
	ActionDeleted = "deleted"
)

// ActionDeletedParms represents the parameters for the deleted action.
type ActionDeletedParms struct {
	UserID uuid.UUID
}

// String returns a string representation of the action parameters.
func (act *ActionDeletedParms) String() string {
	return fmt.Sprintf("&EventParamsUpdated{UserID:%v}", act.UserID)
}

// Marshal returns the event parameters encoded as JSON.
func (act *ActionDeletedParms) Marshal() ([]byte, error) {
	return json.Marshal(act)
}

// ActionDeletedData constructs the data for the deleted action.
func ActionDeletedData(userID uuid.UUID) delegate.Data {
	params := ActionDeletedParms{
		UserID: userID,
	}

	rawParams, err := params.Marshal()
	if err != nil {
		panic(err)
	}

	return delegate.Data{
		Domain:    DomainName,
		Action:    ActionDeleted,
		RawParams: rawParams,
	}
}
