<?php
strncmp('ab', $s, <error descr="Length 3 does not match the 2-character literal.">3</error>);
strncmp('ab', 'abcd', <error descr="Length 2 does not match the 4-character literal.">2</error>);
strncmp($s, 'é', <error descr="Length 1 does not match the 2-character literal.">1</error>);
