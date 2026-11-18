<?php
namespace Negative0 { $x=new \DateTime();$x->modify('+1 day'); }
namespace Negative1 { $x=new \DateTimeImmutable();$x=$x->modify('+1 day'); }
namespace Negative2 { (new \DateTimeImmutable())->format('c'); }
namespace Negative3 { $x->modify('+1 day'); }
namespace Negative4 { $x=new \DateTimeImmutable();$x->{$name}(); }
