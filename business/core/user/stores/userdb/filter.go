package userdb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ardanlabs/service/business/core/user"
)

func (s *Store) applyFilter(filter user.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["user_id"] = *filter.ID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.Name != nil {
		data["name"] = fmt.Sprintf("%%%s%%", *filter.Name)
		wc = append(wc, "name LIKE :name")
	}

	if filter.Email != nil {
		data["email"] = (*filter.Email).String()
		wc = append(wc, "email = :email")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
