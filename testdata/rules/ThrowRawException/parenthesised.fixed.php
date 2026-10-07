<?php
class Silent extends \InvalidArgumentException {}
function pick(int $mode) {
    if ($mode === 1) {
        throw (new \RuntimeException('bad mode'));
    }
    if ($mode === 2) {
        throw ((new Silent()));
    }
    throw (new Silent('unsupported'));
}
