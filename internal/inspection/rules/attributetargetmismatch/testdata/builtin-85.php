<?php
// PHP 8.5 adds NoDiscard, Deprecated on more targets and DelayedTargetValidation.
<error descr="Use the attribute on a permitted target.">#[\NoDiscard]</error> class X {}
class C { #[\NoDiscard] public function m(): int { return 1; } }
#[\Deprecated] const A = 1;
#[\Deprecated] trait T {}
// Among class-likes Deprecated applies to traits alone.
<error descr="Use the attribute on a permitted target.">#[\Deprecated]</error> class Retired {}
<error descr="Use the attribute on a permitted target.">#[\Deprecated]</error> interface Shape {}
<error descr="Use the attribute on a permitted target.">#[\Deprecated]</error> enum Kind {}
// DelayedTargetValidation postpones the check of PHP's own attributes.
#[\DelayedTargetValidation] #[\NoDiscard] class Y {}
#[\DelayedTargetValidation]
#[\Override]
class Z {}
