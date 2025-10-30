package domain

import (
	"errors"
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
