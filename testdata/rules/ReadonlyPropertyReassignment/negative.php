<?php
namespace Negative0 { class P {public int $p;function f(){$this->p=1;$this->p=2;}} }
namespace Negative1 { class P{public readonly int $p;function f(){$this->p=1;}} }
namespace Negative2 { class P{public readonly int $p;function f(){unknown();$this->p=2;}} }
namespace Negative3 { class P{public readonly int $p;function __clone(){$this->p=1;$this->p=2;}} }
namespace Negative4 { $a=1; }
namespace Negative5 { $x->p=1; }
namespace Negative6 { class P{public readonly int $p;function f(){if($q){$this->p=1;}$this->p=2;}} }
namespace Negative7 { class P{public readonly int $p;function f(){$this->{$name}=1;}} }
namespace Negative8 { class P{public readonly int $p;function f(){$this->p=&$x;$this->p=2;}} }
namespace Negative9 { class P{public readonly int $p;function f(){for($this->p=1;;){}}} }
namespace Negative10 { class P{public readonly int $p;function f(){if($q)$this->p=1;}} }

namespace Audit0 { class P {public readonly int $p; static function run() {$o=new P; $o->p=1; $o=new P; $o->p=2;}} }

namespace Audit1 { class P{public readonly int $p;function f(){$this->p=resetProperty($this);$this->p=2;}} }

namespace Audit2 { class P{public readonly int $p;function f(){make()->p=1;make()->p=2;}} }

namespace Audit3 { class P{public readonly int $p;function f(){$this->p=1;echo 'ok';$this->p=2;}} }

namespace Audit4 { class P{public readonly int $p;function f(){$x=1;$this->p=2;}} }

namespace AuditSecond0 { class P{public readonly int $p;static function make():self{return new self;}function f(){self::make()->p=1;self::make()->p=2;}} }
