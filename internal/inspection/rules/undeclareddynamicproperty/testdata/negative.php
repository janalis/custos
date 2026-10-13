<?php
namespace Negative0 { $x->p=1; }
namespace Negative1 { class P {public $p;} (new P())->p=1; }
namespace Negative2 { class P {function __set($n,$v){}} (new P())->p=1; }
namespace Negative3 { #[\AllowDynamicProperties] class P{} (new P())->p=1; }
namespace Negative4 { (new \stdClass())->p=1; }
namespace Negative5 { class P extends Missing{} (new P())->p=1; }
namespace Negative6 { $a=1; }
namespace Negative7 { class P{} (new P())->{$name}=1; }
namespace Negative8 { (new Missing())->p=1; }
namespace Negative9 { class P{}(new P())->{$name}=1; }
namespace Negative10 { class P extends \stdClass{}(new P())->p=1; }

namespace AuditSecond0 { class Base{}class Child extends Base{public $p;}function put(Base $o){$o->p=1;} }

namespace AuditSecond1 { class Base{}class Child extends Base{public $p;}function put(Base $o){if(!property_exists($o,'p')){throw new \Exception;}$o->p=1;} }

namespace AuditSecond3 { final class P{} function put(P $o){if(!property_exists($o,'p')){throw new \Exception;}$o->p=1;} }

namespace AuditSecond4 { final class P{} function put(P $o){if(!property_exists($o,'p')){return;}$o->p=1;} }

namespace AuditSecond5 { class P{}$o=getObject();$o->p=1; }

namespace AuditSecond6 { class P{static function put(){$o=new static;$o->p=1;}} }

namespace AuditSecond7 { class P{}function make():P{return new P;} $o=make();$o->p=1; }

namespace AuditSecond8 { class P{}$o=new P;echo 'go';$o->p=1; }

namespace AuditSecond9 { class P{}$o=new P;$other=new P;$o->p=1; }

namespace AuditSecond10 { class P{}function put(P $o){$o->p=1;} }
