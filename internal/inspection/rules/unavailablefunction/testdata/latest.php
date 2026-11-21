<?php
// Every function PHP declares now is available at the latest version.
str_contains('haystack', 'needle');
array_first([]);
// Removed functions are still reported there.
<error descr="each() was removed in PHP 8.0; the project targets PHP 8.5.">each</error>($items);
