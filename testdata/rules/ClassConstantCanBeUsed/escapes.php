<?php

namespace app\events;

class UserEvent {}

$names = [
    // \e is the ESC byte and \x41 is "A": not the written class names
    "app\events\UserEvent",
    "app\x41bc\Reader",
    <weak_warning descr="Use \app\events\UserEvent::class instead of the class name string.">"app\\events\\UserEvent"</weak_warning>,
    // \x without a hex digit stays as written
    <weak_warning descr="Use \app\xml\Reader::class instead of the class name string.">"app\xml\Reader"</weak_warning>,
];

namespace app\xml;
class Reader {}

namespace app\Abc;
class Reader {}
