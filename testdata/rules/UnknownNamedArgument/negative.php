<?php
namespace Negative0 { function f($x){}f(x:1); }
namespace Negative1 { function f(...$args){}f(anything:1); }
namespace Negative2 { unknown(x:1); }
namespace Negative3 { class C{function f($x){}} (new C())->f(x:1); }
namespace Negative4 { class C{static function f($x){}} C::f(x:1); }
namespace Negative5 { strlen(string:'x'); }
namespace Negative6 { strlen(...); }
namespace Negative7 { function f(){}f(); }
