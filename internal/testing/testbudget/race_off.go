//go:build !race

package testbudget

// raceFactor scales budgets in a race-detector build (see race_on.go).
const raceFactor = 1
