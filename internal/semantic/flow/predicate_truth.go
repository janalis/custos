package flow

const (
	falsyTruth uint8 = 1 << iota
	truthyTruth
)

func scalarTruth(value constantScalar) uint8 {
	if value.truth() {
		return truthyTruth
	}
	return falsyTruth
}

func truthOutcomes(value Value) uint8 {
	if !value.Complete || value.replaced || value.Invalidated != 0 {
		return 0
	}
	if value.truth != 0 {
		return value.truth
	}
	if scalar, known := knownScalar(value); known {
		return scalarTruth(scalar)
	}
	return 0
}

func predicateTruth(value Value) (bool, bool) {
	outcomes := truthOutcomes(value)
	return outcomes == truthyTruth, outcomes == falsyTruth || outcomes == truthyTruth
}
