package usersummary

import (
	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
type QueryFilter struct {
	UserID   *uuid.UUID `validate:"omitempty,uuid4"`
	UserName *string    `validate:"omitempty,min=3"`
}

// WithUserID sets the ID field of the QueryFilter value.
func (qf *QueryFilter) WithUserID(userID uuid.UUID) {
	qf.UserID = &userID
}

// WithUserName sets the UserName field of the QueryFilter value.
func (qf *QueryFilter) WithUserName(userName string) {
	qf.UserName = &userName
}
