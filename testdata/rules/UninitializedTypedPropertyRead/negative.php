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
