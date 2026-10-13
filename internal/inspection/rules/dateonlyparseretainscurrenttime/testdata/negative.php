<?php
namespace Negative0 { \DateTimeImmutable::createFromFormat('!Y-m-d',$d); }
namespace Negative1 { \DateTimeImmutable::createFromFormat('Y-m-d|',$d); }
namespace Negative2 { \DateTimeImmutable::createFromFormat('Y-m-d H:i:s',$d); }
namespace Negative3 { \DateTimeImmutable::createFromFormat('H:i:s',$d); }
namespace Negative4 { \DateTimeImmutable::createFromFormat($format,$d); }
namespace Negative5 { strlen('x'); }
namespace Negative6 { Unknown::createFromFormat('Y-m-d',$date); }

namespace Audit0 { \DateTimeImmutable::createFromFormat('\\Y-\\m-\\d','Y-m-d'); }

namespace Review0 { $a = \DateTimeImmutable::createFromFormat('Y-m-d', $in)->setTime(0, 0); }
namespace Review1 { $a = \DateTimeImmutable::createFromFormat('Y-m-d', $in)?->setTime(0, 0); }
namespace Review2 { $a = (\DateTime::createFromFormat('d/m', $in))->SETTIME(12, 30); }
