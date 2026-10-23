//go:build race

package testbudget

// raceFactor scales budgets in a race-detector build, whose memory
// instrumentation slows the measured work several times while leaving
// calibrate's arithmetic loop nearly as fast.
const raceFactor = 10
