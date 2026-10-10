<?php
namespace Negative0 { $d=new \DateTimeImmutable('2025-03-30 00:00:00',new \DateTimeZone('UTC'));$d->getTimestamp()+86400; }
namespace Negative1 { $d=new \DateTimeImmutable('2025-03-30 12:00:00',new \DateTimeZone('Europe/Paris'));$d->getTimestamp()+86400; }
namespace Negative2 { $ts+86400; }
namespace Negative3 { $d=new \DateTimeImmutable();$d->getTimestamp()+3600; }
namespace Negative4 { $d=new \DateTimeImmutable('2025-03-30 00:00:00');$d->getTimestamp()+86400; }
namespace Negative5 { $d->getTimestamp()-86400; }
namespace Negative6 { 86400+$x; }
namespace Negative7 { $x=new \DateTimeImmutable('2025-01-01 00:00:00',new \DateTimeZone('Unknown'));$x->getTimestamp()+86400; }
namespace Negative8 { $x=new \DateTimeImmutable('2025-99-99 00:00:00',new \DateTimeZone('Europe/Paris'));$x->getTimestamp()+86400; }
namespace Negative9 { $x=new \DateTimeImmutable('2025-01-01 00:00:00',new \DateTimeZone($zone));$x->getTimestamp()+86400; }
namespace Negative10 { $x=new \DateTimeImmutable($date,new \DateTimeZone('Europe/Paris'));$x->getTimestamp()+86400; }
namespace Negative11 { $x=new \DateTimeImmutable();$x->format('c')+86400; }
namespace Negative12 { $x=\DateTimeImmutable::createFromFormat('Y-m-d',$date);$x->getTimestamp()+86400; }

namespace Audit0 { $d=new \DateTime('2025-03-30 00:00:00',new \DateTimeZone('Europe/Paris'));$d->setTimezone(new \DateTimeZone('UTC'));$next=$d->getTimestamp()+86400; }
