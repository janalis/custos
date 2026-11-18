<?php
namespace Negative0 { \DateTimeImmutable::createFromFormat('!Y-m-d','2025-02-28'); }
namespace Negative1 { \DateTimeImmutable::createFromFormat($f,$d); }
namespace Negative2 { \DateTimeImmutable::createFromFormat('!Y-m-d','nonsense'); }
namespace Negative3 { \DateTimeImmutable::createFromFormat('!Y-m-d','2025-02-31');$errors=\DateTimeImmutable::getLastErrors();if($errors){reject();} }
namespace Negative4 { Unknown::createFromFormat('Y-m-d','2025-02-31'); }
namespace Negative5 { \DateTimeImmutable::createFromFormat('!Y-m-d',$unknown); }

namespace Audit0 { \DateTimeImmutable::createFromFormat('Y-m-d','xxxx-xx-xx'); }

namespace Review0 { function parse() { $d = \DateTimeImmutable::createFromFormat('Y-m-d', '2025-02-31'); $e = \date_get_last_errors(); if ($d === false || ($e && ($e['warning_count'] || $e['error_count']))) { return null; } return $d; } }
namespace Review1 { function parse() { $d = \DateTime::createFromFormat('!Y-m-d', '2025-04-31'); $e = date_get_last_errors(); return $e ? null : $d; } }
