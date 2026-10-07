<?php
strncmp($s, 'abc', 3);
strncmp($s, 'abc');
strncmp($s, 'abc', 3, 4);
strncmp($s, "a$b", 2);
strncmp($s, 'abc', 4.0);
strncmp($s, 'abc', LEN);
$o->strncmp($s, 'abc', 5);
strncmp($s, "\x41\101\u{1F600}", 6);
