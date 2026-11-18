<?php
namespace Negative0 { (new \DateInterval('P1D'))->format('%d'); }
namespace Negative1 { $a=new \DateTimeImmutable();$b=new \DateTimeImmutable();$a->diff($b)->format('%a'); }
namespace Negative2 { $x->format('%d'); }
namespace Negative3 { $a=new \DateTimeImmutable();$a->format('%d'); }
namespace Negative4 { class P{function make(){return new \DateInterval('P1D');}}(new P())->make()->format('%d'); }

namespace Audit0 { $a=new \DateTimeImmutable;$b=new \DateTimeImmutable;$i=$a->diff($b);$i->d=0;echo $i->format('%d'); }
