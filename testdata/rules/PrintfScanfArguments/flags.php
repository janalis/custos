<?php
// Flags may come in any order and repeat.
echo sprintf('%0+5d', 42);
echo sprintf('%-+6d|', 42);
echo sprintf('%+ 05d', 4);
echo sprintf("%'x-5d|", 4);
echo sprintf('%\'x-5d|', 4);
echo sprintf('%-010d|', 42);
echo sprintf('%.f', 1.5);
echo <error descr="This call needs 2 argument(s) in total.">sprintf</error>('%0+5d');
echo <error descr="This call needs 2 argument(s) in total.">sprintf</error>('%-+6d|', 1, 2);
// A flag run still needs a conversion character.
echo sprintf(<error descr="Malformed format string.">'%-+'</error>, 4);
echo sprintf(<error descr="Malformed format string.">"%'"</error>, 4);
echo sprintf(<error descr="Malformed format string.">'%lld'</error>, 4);
