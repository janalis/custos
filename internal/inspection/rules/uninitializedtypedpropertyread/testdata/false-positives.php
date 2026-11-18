<?php
namespace Negative0 { class P{public int $p=1;} echo (new P())->p; }
namespace Negative1 { class P{public int $p;function __construct(){$this->p=1;}} echo (new P())->p; }
namespace Negative2 { class P{public int $p;} isset((new P())->p); }
namespace Negative3 { class P{public int $p;} empty((new P())->p); }
namespace Negative4 { class P{public int $p;} (new P())->p=1; }
namespace Negative5 { echo $unknown->p; }
namespace Negative6 { class P{public $p;}echo (new P())->p; }
namespace Negative7 { class P{public int $p;}echo (new P())->{$name}; }
namespace Negative8 { class P{public int $p;}$p=new P();$p->p=1;echo $p->p; }
namespace Negative9 { class P{public int $p;}$p=new P();echo $p->p; }

namespace Audit0 { class P{public int $p;}echo (new P)->p ?? 1; }

namespace AuditThird0 { class P extends Unknown{public int $p;}echo (new P)->p; }

namespace Review0 { class Conf{public array $opts;public ?Conf $next;public int $n;}
    isset((new Conf)->opts['k']);
    isset((new Conf)->next->n);
    echo (new Conf)->opts['k'] ?? 'fallback';
    echo (new Conf)->next->n ?? 'fallback';
    echo empty((new Conf)->opts['k']);
    echo (new Conf)->next?->n ?? 'fallback';
    echo ((new Conf)->opts)['k'] ?? 'fallback';
}
namespace Review1 {
    class Base{public int $v;public function peek(){return (new static)->v;}}
    class Child extends Base{public function __construct(){$this->v=7;}}
}
