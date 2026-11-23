<?php
// PHP 8.6: a readonly property with a default value is initialized when
// the object is created; no later write may change it.
namespace Final0 { final class P { public readonly int $p = 1; function __construct() { <error descr="Initialize this readonly property only once.">$this->p</error> = 2; } } }
namespace Final1 { final class P { public readonly int $p = 1; function f(int $n) { <error descr="Initialize this readonly property only once.">$this->p</error> += $n; } } }
namespace Private0 { class P { private readonly int $p = 1; function __construct() { <error descr="Initialize this readonly property only once.">$this->p</error> = 2; } } }
namespace ReadonlyClass0 { final readonly class P { public int $p = 1; function __construct() { <error descr="Initialize this readonly property only once.">$this->p</error> = 2; } } }
namespace Other0 { final class P { public readonly int $p = 1; static function make(): self { $o = new self(); <error descr="Initialize this readonly property only once.">$o->p</error> = 2; return $o; } } }

// No report.
// A subclass may redeclare the property without the default.
namespace Open0 { class P { public readonly int $p = 1; function __construct() { $this->p = 2; } } }
namespace Open1 { class P { protected readonly int $p = 1; function __construct() { $this->p = 2; } } }
// A private property is another one in a subclass, which may write the name.
namespace Private1 { class B { private readonly int $y = 1; } class C extends B { function __construct() { $this->y = 5; } } }
// __clone may reinitialize it.
namespace Clone0 { final class P { public readonly int $p = 1; function __clone() { $this->p = 2; } } }
// A promoted parameter's default is no property default.
namespace Promoted0 { final class P { function __construct(public readonly int $p = 1) { $this->p = 2; } } }
// No default.
namespace NoDefault0 { final class P { public readonly int $p; function __construct() { $this->p = 2; } } }
// `??=` writes only a null value.
namespace Coalesce0 { final class P { public readonly ?int $p = null; function f() { $this->p ??= 2; } } }
