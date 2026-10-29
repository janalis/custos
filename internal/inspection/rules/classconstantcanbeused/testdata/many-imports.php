<?php

namespace Lib\Tools;
class Hammer {}
class Saw {}
class Drill {}
class Level {}

namespace App;

use Exception;

$list = [
    <weak_warning descr="Use \Lib\Tools\Hammer::class instead of the class name string.">'Lib\Tools\Hammer'</weak_warning>,
    <weak_warning descr="Use \Lib\Tools\Saw::class instead of the class name string.">'Lib\Tools\Saw'</weak_warning>,
    <weak_warning descr="Use \Lib\Tools\Drill::class instead of the class name string.">'Lib\Tools\Drill'</weak_warning>,
    <weak_warning descr="Use \Lib\Tools\Level::class instead of the class name string.">'Lib\Tools\Level'</weak_warning>,
    <weak_warning descr="Use \Lib\Tools\Drill::class instead of the class name string.">'Lib\Tools\Drill'</weak_warning>,
];

namespace Site;

$tools = [
    <weak_warning descr="Use \Lib\Tools\Hammer::class instead of the class name string.">'Lib\Tools\Hammer'</weak_warning>,
    <weak_warning descr="Use \Lib\Tools\Saw::class instead of the class name string.">'Lib\Tools\Saw'</weak_warning>,
    <weak_warning descr="Use \Lib\Tools\Drill::class instead of the class name string.">'Lib\Tools\Drill'</weak_warning>,
];
