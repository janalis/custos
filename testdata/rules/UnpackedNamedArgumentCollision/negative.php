<?php
namespace ArrayOverwrite {
function send($id, $other = 0) {}
send(...['id' => 1, 'id' => 2]);
send(...[0 => 1, 0 => 2], other: 3);
send(...['0' => 1, 0 => 2], other: 3);
send(...[$unknown => 1], id: 3);
}
namespace Negative0 { function f($x,$y){} f(...['x'=>1],y:2); }
namespace Negative1 { function f($x){}f(...$args,x:2); }
namespace Negative2 { function f($x){}f(x:1); }
namespace Negative3 { function f($x){}f(...); }
namespace Negative4 { function f($x){}f(...[...$args],x:2); }
namespace Negative5 { class C{static function f($x){}}C::f(...['x'=>1]); }
namespace Negative6 { unknown(...['x'=>1],x:2); }
namespace Negative7 { function f(){}f(...[1]); }
namespace Negative8 { function f(){}f(1); }
namespace Negative9 { function f($x){}f(...[1]); }
namespace Negative10 { function f($x){}f(1); }
namespace Negative11 { class C{function f($x){}}(new C())->f(...['x'=>1]); }

namespace AuditIndex {function f($x,$y,$z){}f(...[-2=>1,2,-1=>3],z:4);}
