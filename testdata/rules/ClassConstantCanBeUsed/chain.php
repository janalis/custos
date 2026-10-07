<?php

namespace App;

use Exception;

class Kit {}

namespace Lib\Tools;
class Hammer {}
class Saw {}

namespace App;

$list = [
    <weak_warning descr="Use \Lib\Tools\Hammer::class instead of the class name string.">'Lib\Tools\Hammer'</weak_warning>,
    <weak_warning descr="Use \Lib\Tools\Saw::class instead of the class name string.">'Lib\Tools\Saw'</weak_warning>,
    <weak_warning descr="Use \Lib\Tools\Hammer::class instead of the class name string (::class has no leading backslash).">'\Lib\Tools\Hammer'</weak_warning>,
];
