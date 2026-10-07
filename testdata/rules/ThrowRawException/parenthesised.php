<?php
class Silent extends \InvalidArgumentException {}
function pick(int $mode) {
    if ($mode === 1) {
        throw (new <weak_warning descr="Throw a more specific exception class than \Exception.">\Exception</weak_warning>('bad mode'));
    }
    if ($mode === 2) {
        throw ((<weak_warning descr="Pass a message when throwing this exception.">new Silent()</weak_warning>));
    }
    throw (new Silent('unsupported'));
}
