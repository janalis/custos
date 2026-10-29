<?php

namespace Lib\Tools;
class Hammer {}
class Saw {}
class Drill {}
class Level {}

namespace App;

use Exception;
use Lib\Tools\Hammer;
use Lib\Tools\Saw;

$list = [
    Hammer::class,
    Saw::class,
    'Lib\Tools\Drill',
    'Lib\Tools\Level',
    'Lib\Tools\Drill',
];

namespace Site;

use Lib\Tools\Hammer;

$tools = [
    Hammer::class,
    'Lib\Tools\Saw',
    'Lib\Tools\Drill',
];
