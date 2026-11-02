package common

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

// Date is a calendar date with no time component, the Go analogue of Kotlin's LocalDate. It
// serializes as "YYYY-MM-DD" (matching Jackson) and stores in a Postgres DATE column.
type Date struct {
	time.Time
}

// NewDate builds a Date from year/month/day (UTC midnight).
func NewDate(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format(dateLayout) + `"`), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date %q, expected YYYY-MM-DD", s)
	}
	d.Time = t
	return nil
}

// String renders the date or "null" — used when folding into the snapshot version material so it
// matches Kotlin's listOf(...).joinToString rendering of a nullable LocalDate.
func (d *Date) String() string {
	if d == nil {
		return "null"
	}
	return d.Format(dateLayout)
}

func (d Date) Value() (driver.Value, error) {
	return d.Time.Format(dateLayout), nil
}

func (d *Date) Scan(v any) error {
	switch t := v.(type) {
	case time.Time:
		d.Time = t
		return nil
	case string:
		parsed, err := time.Parse(dateLayout, t)
		if err != nil {
			return err
		}
		d.Time = parsed
		return nil
	case nil:
		return nil
	}
	return fmt.Errorf("cannot scan %T into Date", v)
}

// GormDataType maps Date to a DATE column.
func (Date) GormDataType() string { return "date" }
