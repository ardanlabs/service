package darwin

import (
	"bufio"
	"strconv"
	"strings"
)

// ParseMigrations takes a string that represents a text formatted set
// of migrations and parse them for use.
func ParseMigrations(s string) []Migration {
	var migs []Migration

	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(bufio.ScanLines)

	var mig Migration
	var script string
	for scanner.Scan() {
		v := scanner.Text()
		lower := strings.ToLower(v)
		switch {
		case (len(v) >= 5 && lower[:5] == "--ver") || (len(v) >= 6 && lower[:6] == "-- ver"):
			script = strings.TrimSpace(script)
			script = strings.TrimRight(script, "\n")

			mig.Script = script
			migs = append(migs, mig)

			mig = Migration{}
			script = ""

			parts := strings.Split(v, ":")
			if len(parts) != 2 {
				return nil
			}

			f, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				return nil
			}
			mig.Version = f

		case (len(v) >= 5 && lower[:5] == "--des") || (len(v) >= 6 && lower[:6] == "-- des"):
			parts := strings.Split(v, ":")
			if len(parts) != 2 {
				return nil
			}

			mig.Description = strings.TrimSpace(parts[1])

		default:
			script += v + "\n"
		}
	}

	script = strings.TrimSpace(script)
	script = strings.TrimRight(script, "\n")
	mig.Script = script
	migs = append(migs, mig)

	return migs[1:]
}
