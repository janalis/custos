<?php
// PHP 8.6: Override also applies to class constants and enum cases.
namespace Missing0 { class C { <error descr="Apply Override to an inherited member.">#[\Override]</error> const X = 1; } }
namespace Missing1 { class P { private const X = 1; } class C extends P { <error descr="Apply Override to an inherited member.">#[\Override]</error> const X = 2; } }
namespace Missing2 { class P { const X = 1; } class C extends P { <error descr="Apply Override to an inherited member.">#[\Override]</error> const X = 1, Y = 2; } }
namespace Missing3 { enum E { <error descr="Apply Override to an inherited member.">#[\Override]</error> case A; } }
namespace Missing4 { class P { public function X() {} } class C extends P { <error descr="Apply Override to an inherited member.">#[\Override]</error> const X = 1; } }

// No report.
namespace Found0 { class P { const X = 1; } class C extends P { #[\Override] const X = 2; } }
namespace Found1 { interface I { const X = 1; } class C implements I { #[\Override] const X = 2; } }
namespace Found2 { interface I { const X = 1; } interface J extends I { #[\Override] const X = 2; } }
namespace Found3 { interface I { const A = 1; } enum E implements I { #[\Override] case A; } }
namespace Found4 { interface I { const X = 1; } enum E implements I { #[\Override] const X = 2; } }
namespace Found5 { class P { const X = 1, Y = 2; } class C extends P { #[\Override] const X = 3, Y = 4; } }
namespace Found6 { class G { protected const X = 1; } class P extends G {} class C extends P { #[\Override] const X = 2; } }
namespace Found7 { class P { const X = 1; } $o = new class extends P { #[\Override] const X = 3; }; }
namespace Unknown0 { class C extends Missing { #[\Override] const X = 1; } }
namespace Trait0 { trait T { #[\Override] const X = 1; } }
// Another rule reports Override on a parameter.
namespace Param0 { class C { function m(#[\Override] $x) {} } }
