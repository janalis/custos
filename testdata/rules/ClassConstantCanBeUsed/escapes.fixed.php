<?php

namespace app\events;

use app\xml\Reader;

class UserEvent {}

$names = [
    // \e is the ESC byte and \x41 is "A": not the written class names
    "app\events\UserEvent",
    "app\x41bc\Reader",
    \app\events\UserEvent::class,
    // \x without a hex digit stays as written
    Reader::class,
];

namespace app\xml;
class Reader {}

namespace app\Abc;
class Reader {}
