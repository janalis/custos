<?php
class QuietFailure extends \LogicException {
    function __construct()
    {
        parent::__construct("quiet", 7);
    }
}
class PresetFailure extends \DomainException {
    protected $message = 'preset text';
}

function guard($n, $e, $cls) {
    if ($n === 3) {
        throw new QuietFailure();
    }
    if ($n === 4) {
        throw new PresetFailure();
    }
    if ($n === 5) {
        throw $e;
    }
    if ($n === 6) {
        throw new $cls();
    }
    if ($n === 7) {
        throw new UnknownThing();
    }
    throw new LengthException('too long');
}
