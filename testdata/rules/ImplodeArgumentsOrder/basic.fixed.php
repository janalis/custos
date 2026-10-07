<?php

$csv   = implode(';', $cells);
$path  = \implode("/", $segments);
$line  = implode("\n{$eol}", get_rows());
$keys  = implode(',', ['a', 'b']);
