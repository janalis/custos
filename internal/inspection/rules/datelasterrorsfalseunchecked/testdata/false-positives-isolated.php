<?php
namespace Negative0 { $e=\DateTimeImmutable::getLastErrors();if($e){echo $e['warning_count'];} }
namespace Negative1 { $e=\DateTimeImmutable::getLastErrors();if($e!==false){echo $e['warning_count'];} }
namespace Negative2 { $e=\DateTimeImmutable::getLastErrors()?:[];echo $e['warning_count']; }
namespace Negative3 { echo $x['warning_count']; }
namespace Negative4 { $e=\DateTimeImmutable::getLastErrors();if($x!==false){echo 1;} }
namespace Negative5 { $e=\DateTimeImmutable::getLastErrors();if(false!==$e){echo $e['warning_count'];} }
namespace Negative6 { function f(){$e=\DateTimeImmutable::getLastErrors();if($e===false){return;}echo $e['warning_count'];} }
