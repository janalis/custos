<?php
declare(strict_types=1);

// Before PHP 8.0, PHP's own functions do not declare their parameter types
// as the stubs write them; functions of the project are still checked.
function charge(int $cents) {}

strlen(42);
str_repeat('-', '3');
charge(<error descr="Pass int to $cents of charge(), not string.">'1200'</error>);
