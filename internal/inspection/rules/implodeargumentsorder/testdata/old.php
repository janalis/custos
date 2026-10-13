<?php

$a = <weak_warning descr="Pass the separator as the first argument of implode().">implode($parts, <<<EOT
, {$sep}
EOT
)</weak_warning>;
$b = <weak_warning descr="Pass the separator as the first argument of implode().">implode($parts, <<<'EOT'
-
EOT
)</weak_warning>;
$c = <weak_warning descr="Pass the separator as the first argument of implode().">implode($parts, ',')</weak_warning>;
$d = <weak_warning descr="Pass the separator as the first argument of implode().">implode($head + <<<EOT
tail
EOT
, ',')</weak_warning>;
