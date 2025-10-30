package domain

import (
	"errors"
	"fmt"
	"time"
)

// DateRange représente une période temporelle avec validation
// DESIGN PATTERN: Value Object (DDD)
//   - Immutable: pas de setters, valeurs fixées à la création
//   - Validation dans le constructeur (NewDateRangeFromDays)
//   - Égalité basée sur les valeurs, pas l'identité
//
// MÉMOIRE: Taille de la struct
//   - start: time.Time = 24 bytes (wall: 8b, ext: 8b, loc: 8b pointer)
//   - end: time.Time = 24 bytes
//   - TOTAL: 48 bytes
//
// STACK vs HEAP: Dépend du contexte
//   - Si retourné par valeur (DateRange): peut rester sur STACK
//   - Si échappé (passé à interface, stocké dans struct pointeur): va sur HEAP
//   - Go fait l'escape analysis automatiquement
type DateRange struct {
	start time.Time // Minuscule = champ privé (encapsulation)
	end   time.Time
}

// NewDateRangeFromDays crée un DateRange à partir d'un nombre de jours depuis maintenant
// SYNTAXE: Retourne (DateRange, error) par VALEUR, pas pointeur
//   - Value Object pattern: on copie la struct entière
//   - 48 bytes copiés, c'est acceptable (< 10 pointeurs)
//
// MÉMOIRE: DateRange créé sur STACK puis copié vers le caller
//   - Si le caller le stocke dans une struct sur HEAP, sera copié là-bas
//
// PERFORMANCE: time.Now() fait un syscall (lent, ~1-2μs)
//   - time.AddDate fait des calculs de calendrier (complexe)
func NewDateRangeFromDays(days int) (DateRange, error) {
	if days < 0 {
		return DateRange{}, errors.New("days cannot be negative")
	}
	// PERFORMANCE: time.Now() = syscall vers l'horloge système
	//   - Sur Linux: appel VDSO optimisé (pas de context switch)
	//   - ~1-2 microseconde, acceptable pour ce use case
	now := time.Now()
	start := now.AddDate(0, 0, -days)
	// SYNTAXE: DateRange{start: start, end: now}
	//   - Composite literal, crée struct sur STACK (souvent)
	//   - Retour par valeur = copie de 48 bytes
	return DateRange{
		start: start,
		end:   now,
	}, nil
}

// Start retourne la date de début
// SYNTAXE: (dr DateRange) = receiver par VALEUR (pas de pointeur)
//   - DateRange copié lors de l'appel (48 bytes)
//   - Acceptable car lecture seule (immutabilité)
//   - Alternative: (dr *DateRange) économise copie mais moins safe
//
// MÉMOIRE: time.Time retourné par valeur (24 bytes copiés)
func (dr DateRange) Start() time.Time {
	return dr.start
}

// End retourne la date de fin
// PATTERN: Getter pour champ privé (encapsulation)
func (dr DateRange) End() time.Time {
	return dr.end
}

// Duration retourne la durée de la période
// SYNTAXE: time.Duration = int64 représentant des nanosecondes
//   - Ex: 1 seconde = 1_000_000_000 nanoseconds
//
// PERFORMANCE: Sub() fait soustractions arithmétiques (très rapide, ~1ns)
func (dr DateRange) Duration() time.Duration {
	return dr.end.Sub(dr.start)
}

// DaysCount retourne le nombre de jours dans la période
func (dr DateRange) DaysCount() int {
	return int(dr.Duration().Hours() / 24)
}

// Contains vérifie si une date est dans la période
// ALGO: date >= start && date <= end
//   - !Before(start) équivaut à >= start
//   - !After(end) équivaut à <= end
//
// PERFORMANCE: Comparaisons de time.Time très rapides (~2-3ns)
//   - Compare d'abord wall clock, puis monotonic time
func (dr DateRange) Contains(date time.Time) bool {
	return !date.Before(dr.start) && !date.After(dr.end)
}

// Overlaps vérifie si deux périodes se chevauchent
// ALGO: Deux intervalles [a,b] et [c,d] se chevauchent si:
//   - a < d ET c < b
//   - Cas exclus: [1-5] et [6-10] ne se chevauchent pas
//
// PERFORMANCE: 2 comparaisons de time.Time (~4-6ns total)
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
// PATTERN: Value Object equality basée sur les valeurs, pas les références
// SYNTAXE: time.Time.Equal() au lieu de ==
//   - Equal() gère correctement les time zones et monotonic clock
//   - == comparerait les bytes bruts (incorrect pour time.Time)
//
// PERFORMANCE: 2 appels Equal() ~10-20ns total
func (dr DateRange) Equals(other DateRange) bool {
	return dr.start.Equal(other.start) && dr.end.Equal(other.end)
}
