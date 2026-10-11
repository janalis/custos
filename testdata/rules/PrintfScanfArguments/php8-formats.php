<?php
// Additions of PHP 8.0 to the printf family: `h`/`H` and `*` amounts.
echo sprintf('%h', 1.5);
echo sprintf('%H', 1.5);
echo sprintf('%*d', 5, 42);
echo sprintf('%.*f', 2, 3.14159);
echo sprintf('%*.*f', 8, 2, 3.14159);
echo sprintf('%*2$d', 42, 5);
echo sprintf('%1$*2$d', 42, 5);
echo sprintf('%.*2$f', 1.5, 1);
// Each `*` takes one more argument.
echo <error descr="This call needs 3 argument(s) in total.">sprintf</error>('%*d', 42);
echo <error descr="This call needs 4 argument(s) in total.">sprintf</error>('%*.*f', 8, 3.14159);
echo <error descr="This call needs 3 argument(s) in total.">sprintf</error>('%*2$d', 42);
// At most one `*` per amount, and never after digits.
echo sprintf(<error descr="Malformed format string.">'%**d'</error>, 1, 2, 3);
echo sprintf(<error descr="Malformed format string.">'%5*d'</error>, 1, 2);
// printf has no character sets.
echo sprintf(<error descr="Malformed format string.">'%[a-z]'</error>, 'x');
