<?php
namespace Negative0 { date('c',1700000000); }
namespace Negative1 { date('c',$ts/1000); }
namespace Negative2 { date('c',$ts); }
namespace Negative3 { $d->setTimestamp(1730000000000); }
namespace Negative4 { (new \DateTimeImmutable())->format('c'); }
