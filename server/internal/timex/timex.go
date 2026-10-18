package timex

import "time"

const layout = "2006-01-02T15:04:05.000Z"

func ISO(value time.Time) string {
	return value.UTC().Format(layout)
}

func ISOOrNil(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := ISO(*value)
	return &formatted
}
