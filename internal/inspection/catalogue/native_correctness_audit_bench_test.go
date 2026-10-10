package catalogue

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/testing/conformance"
)

func BenchmarkNativeCorrectnessAudit(b *testing.B) {
	for _, id := range []string{
		"SodiumSecretboxNonceReused", "ImagickFrameIndexPastEnd",
		"PdoFetchIntoRetainsSharedRows", "XmlWriterAttributeAfterContent",
		"CurlReadCallbackExceedsRequestedSize", "BufferedWriteFlushedAfterUnlock",
		"NonVoidFunctionFallsThrough", "SqliteNondeterministicFunctionDeclaredDeterministic",
		"SimpleXmlChildValueContainsBareAmpersand", "PdoGroupedFetchReadsRemovedColumn",
		"ParseUrlFailureDereferenced", "PosixAccountLookupFailureDereferenced",
		"RegexReplacementCallbackFallsThrough", "CurlHeaderCallbackMissingByteCount",
		"SysvMessageSerializationMismatch", "TlsHandshakePendingAccepted",
		"FtpPendingTransferAcceptedAsComplete",
	} {
		b.Run(id, func(b *testing.B) {
			marked, err := os.ReadFile(filepath.Join("../../../testdata/rules", id, "basic.php"))
			if err != nil {
				b.Fatal(err)
			}
			src, _, err := conformance.ParseMarkup(marked)
			if err != nil {
				b.Fatal(err)
			}
			e, err := analysis.NewEngine(All(), analysis.Config{Only: []string{id}})
			if err != nil {
				b.Fatal(err)
			}
			file := syntax.Parse("bench.php", src, syntax.Options{})
			e.Analyze(file)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				e.Analyze(file)
			}
		})
	}
}
