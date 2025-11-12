package util

import (
	"errors"
	"time"
)

const DateLayout = "2006-01-02"

// ValidateDate validates that a date string is a valid date
func ValidateDate(dateStr string) (time.Time, error) {
	date, err := time.Parse(DateLayout, dateStr)
	if err != nil {
		return time.Time{}, errors.New("invalid date format")
	}
	if date.IsZero() {
		return time.Time{}, errors.New("invalid date")
	}

	return date, nil
}

// ValidateDateOfBirth validates that DOB is valid and user is at least 18
func ValidateDateOfBirth(dobStr string) error {
	dob, err := ValidateDate(dobStr)
	if err != nil {
		return err
	}
	now := time.Now()

	// check if date is in the future
	if dob.After(now) {
		return errors.New("date of birth cannot be in the future")
	}

	// check if user is at least 18 years old
	if now.Sub(dob) < 0 {
		return errors.New("user must be at least 18 years old")
	}

	return nil
}
