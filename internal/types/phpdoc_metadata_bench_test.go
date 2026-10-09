package types

import "testing"

func BenchmarkPHPDocMetadata(b *testing.B) {
	for name, doc := range map[string]string{
		"nullable_intersection": "?(Reader&Writer)",
		"nullable_shape":        "?array{id: int, label?: string}",
		"conditional_shape":     "(T is int ? array{id: int} : array{id: string})",
		"conditional_generic":   "(T is int ? Collection<string> : Collection<string>)",
		"conditional_callable":  "(T is int ? callable(): int : callable(): int)",
	} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = FromDoc(doc, nil)
			}
		})
	}
}
