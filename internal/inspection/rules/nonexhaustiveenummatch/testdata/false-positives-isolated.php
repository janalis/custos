<?php
namespace Negative0 { enum E{case A;case B;}function f(E $e){return match($e){E::A=>1,E::B=>2};} }
namespace Negative1 { enum E{case A;case B;}function f(E $e){return match($e){E::A=>1,default=>2};} }
namespace Negative2 { match($x){1=>1}; }
namespace Negative3 { class E{}function f(E $e){return match($e){1=>1};} }
namespace Negative4 { enum E{case A;case B;}function f(E $e){return match($e){$x=>1};} }
namespace Negative5 { enum E{case A;case B;}class C{const A=1;}function f(E $e){return match($e){C::A=>1};} }

namespace Audit0 { enum E{case A;case B;} $e=E::A;echo match($e){E::A=>1}; }
