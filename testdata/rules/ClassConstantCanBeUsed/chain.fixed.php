<?php

namespace Lib\Tools;
class Hammer {}
class Saw {}

namespace App;

use Exception;
use Lib\Tools\Hammer;
use Lib\Tools\Saw;

class Kit {}

$list = [
    Hammer::class,
    Saw::class,
    '\Lib\Tools\Hammer',
];
