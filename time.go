package mildtg

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	invalidDTG = "INVALID DTG"
)

const (
	// FullYearFormat is the layout for a full year date-time-group.
	FullYearFormat = "020106Z JAN 2006"

	// ShortYearFormat is the layout for a short year date-time-group.
	ShortYearFormat = "020106Z JAN 06"
)

var (
	// ErrNotEnoughChars is returned when an invalid date-time-group is provided.
	ErrNotEnoughChars = errors.New("date-time-group too short minimum is ddhhmm")

	// ErrInvalidDateTimeGroup is returned when an invalid date-time-group is provided.
	ErrInvalidDateTimeGroup = errors.New("invalid date-time-group")
)

// Time wraps a time.Time to allow for custom
// formatting and parsing of various U.S. military
// date-time-group formats.
type Time struct {
	time.Time
}

// Format returns the date-time-group in the format
func (t Time) Format(layout string) string {
	switch layout {
	case FullYearFormat:
		return t.toString(true)
	case ShortYearFormat:
		return t.toString(false)
	default:
		return t.Time.Format(layout)
	}
}

// String returns the date-time-group in the format
func (t Time) String() string {
	return t.toString(false)
}

// toString returns the date-time-group in the format
// with a long year or a short year.
func (t Time) toString(longYear bool) string {
	if t.IsZero() {
		return invalidDTG
	}

	days := t.Day()
	hours := t.Hour()
	minutes := t.Minute()
	seconds := t.Second()
	month := t.Month()
	year := t.Year()
	tz := t.Location().String()

	b := bytes.NewBuffer(make([]byte, 0, 30))

	// Day
	if days < 10 {
		b.WriteString("0")
	}
	b.WriteString(fmt.Sprintf("%d", days))

	// Hour
	if hours < 10 {
		b.WriteString("0")
	}
	b.WriteString(fmt.Sprintf("%d", hours))

	// Minute
	if minutes < 10 {
		b.WriteString("0")
	}
	b.WriteString(fmt.Sprintf("%d", minutes))

	// Only export seconds if they are not zero
	if seconds != 0 {
		if seconds < 10 {
			b.WriteString("0")
		}
		b.WriteString(fmt.Sprintf("%d", seconds))
	}

	// Timezone
	b.WriteString(tz)

	b.WriteString(" ")

	// Month
	b.WriteString(strings.ToUpper(month.String()[0:3]))

	b.WriteString(" ")

	// Year
	if longYear {
		b.WriteString(fmt.Sprintf("%d", year))
	} else {
		twoDigitYear := year % 100
		if twoDigitYear < 10 {
			b.WriteString("0")
		}
		b.WriteString(fmt.Sprintf("%d", year%100))
	}

	return b.String()
}

// NewTime returns a new Time object.
func NewTime(t time.Time) Time {
	return Time{t}
}

// ParseDTG parses a military date-time-group string in the format
// DDHH[MM]|[MMSS]|(A-Z)[ MMM YY[YY] and returns a Time object.
func ParseDTG(s string) (Time, error) {
	return parseDTGBytes(s)
}

// ParseDTGBytes parses a military date-time-group byte slice in the format
// DDHH[MM]|[MMSS]|(A-Z)[ MMM YY[YY] and returns a Time object.
func parseDTGBytes(s string) (Time, error) {

	// The digitsBeforeChar slice is used to store the digits before any
	// characters in the date-time-group.
	// The slice has enough capacity to store the maximum number of digits
	// before a character in the date-time-group.
	// If there are no characters in the date-time-group, the slice will
	// store the maximum number of digits.
	digitsBeforeChar := make([]byte, 0, 2*4+4)
	maxDigitsBeforeChar := 2*4 + 4

	// The digitsAfterChar slice is used to store the digits after any
	// characters in the date-time-group.
	// This slice will have enough capacity to store a four-digit year.
	digitsAfterChar := make([]byte, 0, 4)
	maxDigitsAfterChar := 4

	// September is the longest month name plus one for the time zone designation.
	chars := make([]byte, 0, 9+1)
	maxChars := 9 + 1

	// The index is used to keep track of the current index in the digits slice.
	digitsBeforeIndex := 0
	digitsAfterIndex := 0
	charIndex := 0
	charsFound := false

	// Iterate over each byte in the byte slice.
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] >= '0' && s[i] <= '9':
			// Digit.
			if !charsFound {
				// If the index is greater than or equal to the length of the slice,
				// return an error.
				if digitsBeforeIndex >= maxDigitsBeforeChar {
					return Time{}, ErrInvalidDateTimeGroup
				}

				digitsBeforeChar = append(digitsBeforeChar, s[i])
				digitsBeforeIndex++

			} else {
				// If the index is greater than or equal to the length of the slice,
				// return an error.
				// This could happen if the method receives a year with more than
				// four digits.
				if digitsAfterIndex >= maxDigitsAfterChar {
					return Time{}, ErrInvalidDateTimeGroup
				}

				digitsAfterChar = append(digitsAfterChar, s[i])
				digitsAfterIndex++
			}

		case s[i] >= 'A' && s[i] <= 'Z' || s[i] >= 'a' && s[i] <= 'z':
			// Character.
			if charIndex >= maxChars {
				return Time{}, ErrInvalidDateTimeGroup
			}

			var char byte
			if s[i] >= 'a' && s[i] <= 'z' {
				char = s[i] - 32
			} else {
				char = s[i]
			}

			chars = append(chars, char)
			charIndex++

			// Set the charsFound flag to true.
			charsFound = true

		case s[i] == ' ':
			// Do nothing.
			continue

		default:
			// Invalid character.
			return Time{}, ErrInvalidDateTimeGroup
		}
	}

	// The digitsBeforeChar slice must have at least six digits
	// and have a zero remainder if len(digitsBeforeChar) is divided by 2.
	if len(digitsBeforeChar) < 6 || len(digitsBeforeChar)%2 != 0 {
		return Time{}, ErrNotEnoughChars
	}

	// The day, hour, and minute are extracted from the digitsBeforeChar slice.
	day := int(digitsBeforeChar[0]-'0')*10 + int(digitsBeforeChar[1]-'0')
	hour := int(digitsBeforeChar[2]-'0')*10 + int(digitsBeforeChar[3]-'0')
	minute := int(digitsBeforeChar[4]-'0')*10 + int(digitsBeforeChar[5]-'0')
	seconds := 0
	year := time.Now().UTC().Year()
	month := time.Now().UTC().Month()
	tz := ZULU

	// Remove the day, hour, and minute from the slice.
	digitsBeforeChar = digitsBeforeChar[6:]

	switch {
	case len(digitsBeforeChar) == 0:
		// Do nothing.
	case len(digitsBeforeChar) == 2:
		// If the length of the remaining digits before the character is two, we
		// can assume these two digits represent the seconds.
		seconds = int(digitsBeforeChar[0]-'0')*10 + int(digitsBeforeChar[1]-'0')
		if seconds > 59 {
			return Time{}, ErrInvalidDateTimeGroup
		}
	case len(digitsBeforeChar) == 4:
		// If the length of the remaining digits before the character is four, we
		// can assume these four digits represent the four-digit year.
		// We do not attempt to parse a two-digit second with a two-digit year.
		year = int(digitsBeforeChar[0]-'0')*1000 + int(digitsBeforeChar[1]-'0')*100 +
			int(digitsBeforeChar[2]-'0')*10 + int(digitsBeforeChar[3]-'0')

	case len(digitsBeforeChar) == 6:
		// If the length of the remaining digits before the character is six, we
		// can assume we have a two-digit seconds and a four-digit year.
		seconds = int(digitsBeforeChar[0]-'0')*10 + int(digitsBeforeChar[1]-'0')
		if seconds > 59 {
			return Time{}, ErrInvalidDateTimeGroup
		}

		year = int(digitsBeforeChar[2]-'0')*1000 + int(digitsBeforeChar[3]-'0')*100 +
			int(digitsBeforeChar[4]-'0')*10 + int(digitsBeforeChar[5]-'0')
	}

	// Parse the month and time zone from the chars slice.
	switch {
	case len(chars) == 0:
		// Do nothing.
	case len(chars) == 1:
		// If the length of the chars slice is one, we can assume this character
		// represents the time zone.
		tzStr := strings.ToUpper(string(chars[0]))
		tzOut, ok := timeZones[rune(tzStr[0])]
		if !ok {
			// Get the local offset.
			offset := time.Now().UTC().Sub(time.Now()).Seconds() / 3600
			tzOut = timeZone{letter: rune(tzStr[0]), offset: int32(offset)}
		}

		tz = tzOut
	case len(chars) == 3:
		// If the length of the chars slice is three, we can assume this represents
		// the three-letter month abbreviation.
		monthStr := string(chars)
		monthOut, ok := months[strings.ToUpper(monthStr)]
		if !ok {
			return Time{}, ErrInvalidMonth
		}

		month = monthOut
	case len(chars) > 3:
		// If the length of the chars slice is greater than three, we either have a
		// time zone, a three-letter month abbreviation, or a full month name or a
		// combination of these.
		//
		// Check if the char string contains a month.
		m, ok := months[strings.ToUpper(string(chars))]
		if !ok {
			// Remove the first character from the chars slice.
			tzStr := strings.ToUpper(string(chars[0]))
			newMonthStr := string(chars[1:])

			// Check if the new month string is a valid month.
			m, ok = months[strings.ToUpper(newMonthStr)]
			if !ok {
				return Time{}, ErrInvalidMonth
			}

			tzOut, tzFound := timeZones[rune(tzStr[0])]
			if !tzFound {
				// Get the local offset.
				offset := time.Now().UTC().Sub(time.Now()).Seconds() / 3600
				tzOut = timeZone{letter: rune(tzStr[0]), offset: int32(offset)}
			}

			tz = tzOut
			month = m
		}

		month = m

	default:
		return Time{}, ErrInvalidDateTimeGroup
	}

	// Check the digitsAfterChar slice.
	if len(digitsAfterChar)%2 != 0 {
		return Time{}, ErrInvalidDateTimeGroup
	}

	// The maximum length of the digitsAfterChar slice is four,
	// and we should not see 1 or 3 digits.
	switch len(digitsAfterChar) {
	case 0:
		// Do nothing.
	case 2:
		// Two-digit year.
		y := int(digitsAfterChar[0]-'0')*10 + int(digitsAfterChar[1]-'0')

		// Determine if the year is in the 21st or 20th century.
		if y < 69 {
			year = 2000 + y
		} else {
			year = 1900 + y
		}
	case 4:
		// Four-digit year.
		y := int(digitsAfterChar[0]-'0')*1000 + int(digitsAfterChar[1]-'0')*100 +
			int(digitsAfterChar[2]-'0')*10 + int(digitsAfterChar[3]-'0')

		year = y
	}

	// Check if the day is valid for the month and year.
	if day > daysInMonth(month, year) || day < 1 {
		return Time{}, ErrInvalidDay
	}

	// Check hours and minutes.
	if hour > 23 || minute > 59 {
		return Time{}, ErrInvalidDateTimeGroup
	}

	t := time.Date(year, month, day, hour, minute, seconds, 0, tz.Location())

	return NewTime(t), nil
}
