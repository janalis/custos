<?php
namespace Lib;

use function Other\count;

function strlen($s) { return 0; }

$a = strlen('x');
$b = count([]);
$c = \strlen('x');
