package auditdb

import (
	"bytes"
	"strings"

	"github.com/ardanlabs/service/business/domain/auditbus"
)

func applyFilter(filter auditbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.Action != nil {
		data["action"] = *filter.Action
		wc = append(wc, "action = :action")
	}

	if filter.Since != nil {
		data["since"] = *filter.Since
		wc = append(wc, "timestamp >= :since")
	}

	if filter.Until != nil {
		data["until"] = *filter.Until
		wc = append(wc, "timestamp <= :until")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
