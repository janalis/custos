<?php
namespace Negative0 { $a->p=1; }
namespace Negative1 { $a->child=2;$b=clone $a;$b->child->p=1; }
namespace Negative2 { $a->child=new \stdClass();$b=$a;$b->child->p=1; }
namespace Negative3 { $a->child=new \stdClass();$b=clone $a;unknown();$b->child->p=1; }
namespace Negative4 { class A{function __clone(){}}$a=new A();$a->child=new \stdClass();$b=clone $a;$b->child->p=1; }
namespace Negative5 { $b->child->p=1; }
namespace Negative6 { $x=1;$a->child=new \stdClass();$z=clone $a;$b->child->p=1; }
namespace Negative7 { $a->child=new \stdClass();$b=clone $a;echo 1;$b->child->p=1; }
namespace Negative8 { $a->different=new \stdClass();$b=clone $a;$b->child->p=1; }
namespace Negative9 { if($q){echo 1;}$b=clone $a;$b->child->p=1; }
namespace Negative10 { echo 1;$b=clone $a;$b->child->p=1; }
namespace Negative11 { $a=new \stdClass();$b=clone $a;$b->child->p=1; }
namespace Negative12 { $a->child=1;$b=clone $a;$b->child->p=1; }
namespace Negative13 { $a=new \stdClass();if($q){echo 1;}$b=clone $a;$b->child->p=1; }
namespace Negative14 {  $a=new \stdClass();echo 1;$b=clone $a;$b->child->p=1; }
namespace Negative15 { $a=new \stdClass();$a->different=new \stdClass();$b=clone $a;$b->child->p=1; }
namespace Negative16 { $a=new \stdClass();$a->child=1;$b=clone $a;$b->child->p=1; }

namespace AuditSecond0 { class P{public function __set($n,$v){} public function __get($n){return new \stdClass;}}$o=new P;$o->child=new \stdClass;$c=clone $o;$c->child->p=1; }

namespace AuditSecond1 { class P{public object $child;}$o=new P;$o->child=new \stdClass;$c=clone($o,['child'=>new \stdClass]);$c->child->p=1; }

namespace AuditThird0 { class P extends Unknown{public object $child;}$o=new P;$o->child=new \stdClass;$c=clone $o;$c->child->p=1; }

namespace AuditThird1 { class P{public object $child;}$o=new P;$o->{$name}=new \stdClass;$c=clone $o;$c->{$name}->p=1; }

namespace AuditThird2 { class P{public object $child{get{return new \stdClass;}set{}}}$o=new P;$o->child=new \stdClass;$c=clone $o;$c->child->p=1; }
