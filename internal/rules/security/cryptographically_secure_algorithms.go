package security

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// cryptographicallySecureAlgorithms reports constants selecting weak
// ciphers or hashes (DES, 3DES, RC2, RC4, MD5, non-AES Rijndael).
type cryptographicallySecureAlgorithms struct{}

func init() { register(cryptographicallySecureAlgorithms{}) }

type weakAlgo struct{ family, alternative string }

const (
	altRijndael = "MCRYPT_RIJNDAEL_128"
	altAES      = "an AES-128-* cipher"
	altBlowfish = "CRYPT_BLOWFISH"
)

var weakAlgos = map[string]weakAlgo{
	"MCRYPT_RIJNDAEL_192":   {"Rijndael with a 192-bit block is not AES", altRijndael},
	"MCRYPT_RIJNDAEL_256":   {"Rijndael with a 256-bit block is not AES", altRijndael + " with a 256-bit key"},
	"MCRYPT_3DES":           {"3DES", altRijndael},
	"MCRYPT_TRIPLEDES":      {"3DES", altRijndael},
	"MCRYPT_DES_COMPAT":     {"DES", altRijndael},
	"MCRYPT_DES":            {"DES", altRijndael},
	"MCRYPT_RC2":            {"RC2", altRijndael},
	"MCRYPT_RC4":            {"RC4", altRijndael},
	"MCRYPT_ARCFOUR":        {"RC4", altRijndael},
	"OPENSSL_CIPHER_3DES":   {"3DES", altAES},
	"OPENSSL_CIPHER_DES":    {"DES", altAES},
	"OPENSSL_CIPHER_RC2_40": {"RC2", altAES},
	"OPENSSL_CIPHER_RC2_64": {"RC2", altAES},
	"CRYPT_MD5":             {"MD5", altBlowfish},
	"CRYPT_STD_DES":         {"DES", altBlowfish},
}

func (cryptographicallySecureAlgorithms) ID() string { return "CryptographicallySecureAlgorithms" }

func (cryptographicallySecureAlgorithms) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KConstFetch}
}

func (cryptographicallySecureAlgorithms) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.ConstFetch)
	if c.Name == nil {
		return
	}
	name := util.LastNamePart(c.Name.Value)
	w, ok := weakAlgos[name]
	if !ok || !cryptoNamesGlobal(ctx, c.Name) || isTestContext(ctx, n) {
		return
	}
	ctx.ReportNode(n, "Weak algorithm selected via "+name+" ("+w.family+"); prefer "+w.alternative+".")
}

// cryptoNamesGlobal reports whether the constant reference resolves to the
// global constant of its last segment (D1): fully qualified, or unqualified
// with PHP's global fallback (no `use const` import and no constant of that
// name declared in the current namespace).
func cryptoNamesGlobal(ctx *analysis.Context, nm *syntax.Name) bool {
	fqn, fallback := ctx.Names().Const(nm.Value, nm.Span().Start)
	if fallback != "" {
		if ctx.Index().Constant(fqn, ctx.PHP) != nil {
			return false
		}
		fqn = fallback
	}
	return !strings.Contains(fqn, `\`)
}

// isTestContext implements D2: a test file path (spec list, deliberately
// narrower than ctx.IsTestFile) or a class whose FQN looks like a test class.
func isTestContext(ctx *analysis.Context, n syntax.Node) bool {
	if util.IsTestPath(ctx.File.Path) {
		return true
	}
	for p := n.Parent(); p != nil; p = p.Parent() {
		cl, ok := p.(*syntax.ClassLike)
		if !ok || cl.Name == nil {
			continue
		}
		fqn := ctx.Names().Class(cl.Name.Value, cl.Name.Span().Start)
		fqn = `\` + fqn
		return strings.HasSuffix(fqn, "Test") || strings.Contains(fqn, `\Tests\`) || strings.Contains(fqn, `\Test\`)
	}
	return false
}
