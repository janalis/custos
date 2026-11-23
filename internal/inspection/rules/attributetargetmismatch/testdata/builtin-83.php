<?php
// PHP's own attributes, with the targets PHP 8.3 declares for them.
<error descr="Use the attribute on a permitted target.">#[\SensitiveParameter]</error> function a() {}
<error descr="Use the attribute on a permitted target.">#[\AllowDynamicProperties]</error> function b() {}
class C { <error descr="Use the attribute on a permitted target.">#[\ReturnTypeWillChange]</error> public $p; }
<error descr="Use the attribute on a permitted target.">#[\Attribute]</error> function d() {}
function e(#[\SensitiveParameter] $x) {}
#[\AllowDynamicProperties] class F {}
// Not yet PHP's own attributes in 8.3: classes the project may declare.
#[\NoDiscard] class G {}
#[\Deprecated] class H {}
