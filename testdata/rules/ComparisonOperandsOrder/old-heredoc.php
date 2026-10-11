<?php
// Before PHP 7.3 the closing marker must stay alone on its line: no fix.
$a = "x";
var_dump(<weak_warning descr="Put the constant operand on the right side of the comparison."><<<EOT
x
EOT
 === $a</weak_warning>);
var_dump(<weak_warning descr="Put the constant operand on the right side of the comparison."><<<'EOT'
x
EOT
 === $a</weak_warning>);
var_dump(<weak_warning descr="Put the constant operand on the right side of the comparison.">null === $a . <<<EOT
$a
EOT</weak_warning>
);
var_dump(<weak_warning descr="Put the constant operand on the right side of the comparison.">'x' === $a</weak_warning>);
