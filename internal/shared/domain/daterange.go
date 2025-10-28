package domain

import (
	"errors"
	"fmt"
	"time"
)

// DateRange représente une période temporelle avec validation
type DateRange struct {
	start time.Time
	end   time.Time
}

// NewDateRange crée une nouvelle instance de DateRange avec validation
func NewDateRange(start, end time.Time) (DateRange, error) {
	if end.Before(start) {
		return DateRange{}, errors.New("end date cannot be before start date")
	}
	return DateRange{
		start: start,
		end:   end,
	}, nil
}

// NewDateRangeFromDays crée un DateRange à partir d'un nombre de jours depuis maintenant
func NewDateRangeFromDays(days int) (DateRange, error) {
	if days < 0 {
		return DateRange{}, errors.New("days cannot be negative")
	}
	now := time.Now()
	start := now.AddDate(0, 0, -days)
	return DateRange{
		start: start,
		end:   now,
	}, nil
}

// Start retourne la date de début
func (dr DateRange) Start() time.Time {
	return dr.start
}

// End retourne la date de fin
func (dr DateRange) End() time.Time {
	return dr.end
}

// Duration retourne la durée de la période
func (dr DateRange) Duration() time.Duration {
	return dr.end.Sub(dr.start)
}

// DaysCount retourne le nombre de jours dans la période
func (dr DateRange) DaysCount() int {
	return int(dr.Duration().Hours() / 24)
}

// Contains vérifie si une date est dans la période
func (dr DateRange) Contains(date time.Time) bool {
	return !date.Before(dr.start) && !date.After(dr.end)
}

// Overlaps vérifie si deux périodes se chevauchent
func (dr DateRange) Overlaps(other DateRange) bool {
	return dr.start.Before(other.end) && other.start.Before(dr.end)
}

// String retourne une représentation textuelle
func (dr DateRange) String() string {
	return fmt.Sprintf("%s to %s (%d days)",
		dr.start.Format("2006-01-02"),
		dr.end.Format("2006-01-02"),
		dr.DaysCount())
}

// Equals vérifie l'égalité entre deux DateRange
func (dr DateRange) Equals(other DateRange) bool {
	return dr.start.Equal(other.start) && dr.end.Equal(other.end)
}
