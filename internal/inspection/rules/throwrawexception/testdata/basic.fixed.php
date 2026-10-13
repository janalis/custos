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
        throw new \RuntimeException("negative $n");
    }
    if ($n === 0) {
        throw new \RuntimeException();
    }
    if ($n === 1) {
        throw new OutOfRangeException;
    }
    if ($n === 2) {
        throw new PlainFailure();
    }
    if ($n === 3) {
        throw new InheritsPreset();
    }
    $x = $n ?? throw new \RuntimeException('expr');
    return $x;
}
