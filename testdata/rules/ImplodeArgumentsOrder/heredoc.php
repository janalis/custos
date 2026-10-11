<?php

$a = <weak_warning descr="Pass the separator as the first argument of implode().">implode($parts, <<<EOT
, 
EOT
)</weak_warning>;
$b = <weak_warning descr="Pass the separator as the first argument of implode().">implode($parts, <<<'EOT'
-
EOT
)</weak_warning>;
