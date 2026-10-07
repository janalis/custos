<?php

namespace App;

use Exception;
use Lib\Tools\Hammer;
use Lib\Tools\Saw;

class Kit {}

namespace Lib\Tools;
class Hammer {}
class Saw {}

namespace App;

$list = [
    Hammer::class,
    Saw::class,
    Hammer::class,
];
