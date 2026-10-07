<?php
class BaseFailure extends \RuntimeException {
    protected $message = 'base failure';
}
class ChildFailure extends BaseFailure {}
class GrandChildFailure extends ChildFailure {}

trait PresetMessage {
    protected $message = 'from trait';
}
class TraitFailure extends \RuntimeException {
    use PresetMessage;
}
class TraitChildFailure extends TraitFailure {}

class BareFailure extends \RuntimeException {}
class BareChildFailure extends BareFailure {}

function raise($n) {
    if ($n === 1) {
        throw new ChildFailure();
    }
    if ($n === 2) {
        throw new GrandChildFailure();
    }
    if ($n === 3) {
        throw new TraitChildFailure();
    }
    if ($n === 4) {
        throw new TraitFailure();      // the class itself uses the trait
    }
    throw <weak_warning descr="Pass a message when throwing this exception.">new BareChildFailure()</weak_warning>;
}
