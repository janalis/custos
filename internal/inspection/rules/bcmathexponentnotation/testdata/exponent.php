<?php
$a = <error descr="Use ordinary decimal notation for BCMath operands.">bcpow('2', '1e2')</error>;
$b = <error descr="Use ordinary decimal notation for BCMath operands.">bcpowmod('2', '1e1', '7')</error>;
$c = <error descr="Use ordinary decimal notation for BCMath operands.">bcpowmod(num: '2', modulus: '7', exponent: '1E1')</error>;
