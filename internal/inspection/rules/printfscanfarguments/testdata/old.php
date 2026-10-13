<?php
// Before PHP 8.0 `h`, `H` and `*` are not part of the printf grammar.
echo sprintf(<error descr="Malformed format string.">'%h'</error>, 1.5);
echo sprintf(<error descr="Malformed format string.">'%H'</error>, 1.5);
echo sprintf(<error descr="Malformed format string.">'%*d'</error>, 5, 42);
echo sprintf(<error descr="Malformed format string.">'%.*f'</error>, 2, 3.14159);
// Flag order and the scanf grammar do not depend on the version.
echo sprintf('%0+5d', 42);
sscanf($line, '%*d %i', $n);
