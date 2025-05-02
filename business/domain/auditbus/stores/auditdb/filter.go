package auditdb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ardanlabs/service/business/domain/auditbus"
)

func applyFilter(filter auditbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ObjID != nil {
		data["obj_id"] = filter.ObjID
		wc = append(wc, "obj_id = :obj_id")
	}

	if filter.ObjDomain != nil {
		data["obj_domain"] = filter.ObjDomain.String()
		wc = append(wc, "obj_domain = :obj_domain")
	}

	if filter.ObjName != nil {
		data["obj_name"] = fmt.Sprintf("%%%s%%", filter.ObjName.String())
		wc = append(wc, "obj_name LIKE :obj_name")
	}

	if filter.ActorID != nil {
		data["actor_id"] = filter.ActorID
		wc = append(wc, "actor_id = :actor_id")
	}

	if filter.Action != nil {
		data["action"] = filter.Action
		wc = append(wc, "action = :action")
	}

	if filter.Since != nil {
		data["since"] = filter.Since
		wc = append(wc, "timestamp >= :since")
	}

	if filter.Until != nil {
		data["until"] = filter.Until
		wc = append(wc, "timestamp <= :until")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
