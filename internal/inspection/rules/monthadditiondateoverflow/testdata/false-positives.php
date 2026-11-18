<?php
namespace Negative0 { (new \DateTimeImmutable('2025-01-28'))->modify('+1 month'); }
namespace Negative1 { (new \DateTimeImmutable($date))->modify('+1 month'); }
namespace Negative2 { (new \DateTimeImmutable('2025-01-31'))->modify('+1 day'); }
namespace Negative3 { (new \DateTimeImmutable('invalid'))->modify('+1 month'); }
namespace Negative4 { $x->modify('+1 month'); }
namespace Negative5 { $x=\DateTimeImmutable::createFromFormat('!Y-m-d',$date);$x->modify('+1 month'); }
namespace Negative6 { (new \DateTimeImmutable('2025-03-28'))->modify('-1 month'); }

namespace Audit0 { $d=new \DateTime('2025-01-31');$d->modify('first day of this month');$d->modify('+1 month'); }
