<?php

$a = implode($parts, <<<EOT
, {$sep}
EOT
);
$b = implode($parts, <<<'EOT'
-
EOT
);
$c = implode(',', $parts);
$d = implode($head + <<<EOT
tail
EOT
, ',');
