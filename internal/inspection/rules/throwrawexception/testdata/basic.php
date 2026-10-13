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
class PlainFailure extends \UnderflowException {}
class InheritsPreset extends PresetFailure {}

function guard($n) {
    if ($n < 0) {
        throw new <weak_warning descr="Throw a more specific exception class than \Exception.">\Exception</weak_warning>("negative $n");
    }
    if ($n === 0) {
        throw new <weak_warning descr="Throw a more specific exception class than \Exception.">Exception</weak_warning>();
    }
    if ($n === 1) {
        throw <weak_warning descr="Pass a message when throwing this exception.">new OutOfRangeException</weak_warning>;
    }
    if ($n === 2) {
        throw <weak_warning descr="Pass a message when throwing this exception.">new PlainFailure()</weak_warning>;
    }
    if ($n === 3) {
        throw new InheritsPreset();
    }
    $x = $n ?? throw new <weak_warning descr="Throw a more specific exception class than \Exception.">Exception</weak_warning>('expr');
    return $x;
}
