<?php
$a = "x";
$r = <weak_warning descr="Put the constant operand on the left side of the comparison.">$a === <<<EOT
x
EOT</weak_warning>;
