package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkExpansionCryptoDBNativeContracts(b *testing.B) {
	for _, tc := range []struct {
		id, src string
		kinds   []syntax.NodeKind
		check   func(*analysis.Context, syntax.Node, string)
	}{
		{"SodiumSecretboxNonceReused", "<?php $k=random_bytes(32);$n=random_bytes(24);sodium_crypto_secretbox('first',$n,$k);sodium_increment($n);sodium_crypto_secretbox('second',$n,$k);", []syntax.NodeKind{syntax.KFuncCall}, CheckSodiumSecretboxNonceReused},
		{"CurlReadCallbackExceedsRequestedSize", "<?php $h=curl_init();curl_setopt($h,CURLOPT_READFUNCTION,function($h,$stream,$size){--$size;return str_repeat('x',$size+1);});", []syntax.NodeKind{syntax.KFuncCall}, CheckCurlReadCallbackExceedsRequestedSize},
		{"HashContextUsedAfterFinalization", "<?php $h=hash_init('sha256');hash_final($h);" + strings.Repeat("hash_update($h,'x');", 100), []syntax.NodeKind{syntax.KFuncCall}, CheckHashContextUsedAfterFinalization},
		{"PdoRepeatedNamedMarkerWithoutEmulation", "<?php $p=new PDO('mysql:host=x');$p->setAttribute(PDO::ATTR_EMULATE_PREPARES,false);" + strings.Repeat("$p->prepare('SELECT :a+:a');", 100), []syntax.NodeKind{syntax.KMethodCall}, CheckPdoRepeatedNamedMarkerWithoutEmulation},
	} {
		b.Run(tc.id, func(b *testing.B) {
			p := expansionCryptoDBProbe{id: tc.id, kinds: tc.kinds, check: func(ctx *analysis.Context, n syntax.Node) { tc.check(ctx, n, "finding") }}
			e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
			if err != nil {
				b.Fatal(err)
			}
			file := syntax.Parse("bench.php", []byte(tc.src), syntax.Options{})
			e.Analyze(file)
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				e.Analyze(file)
			}
		})
	}
}

func BenchmarkExpansionSecretboxDense(b *testing.B) {
	p := expansionCryptoDBProbe{id: "SodiumSecretboxNonceReused", kinds: []syntax.NodeKind{syntax.KFuncCall}, check: func(ctx *analysis.Context, n syntax.Node) { CheckSodiumSecretboxNonceReused(ctx, n, "finding") }}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php $k=random_bytes(32);$n=random_bytes(24);"+strings.Repeat("sodium_crypto_secretbox('same',$n,$k);", 100)), syntax.Options{})
	e.Analyze(file)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}
