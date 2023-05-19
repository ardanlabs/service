package summarydb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ardanlabs/service/business/cview/user/summary"
)

func (s *Store) applyFilter(filter summary.QueryFilter, data map[string]interface{}, buf *bytes.Buffer) {
	var wc []string

	if filter.UserID != nil {
		data["user_id"] = *filter.UserID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.UserName != nil {
		data["user_name"] = fmt.Sprintf("%%%s%%", *filter.UserName)
		wc = append(wc, "user_name LIKE :user_name")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
