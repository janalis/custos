<?php
// Before PHP 7.3 the closing marker must stay alone on its line: no fix.
$a = "x";
var_dump(<<<EOT
x
EOT
 === $a);
var_dump(<<<'EOT'
x
EOT
 === $a);
var_dump(null === $a . <<<EOT
$a
EOT
);
var_dump($a === 'x');
