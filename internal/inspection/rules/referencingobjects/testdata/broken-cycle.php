<?php
// An inheritance cycle: the class counts as overriding itself (D6).
class CycleFirst extends CycleSecond
{
    public function walk(\DOMNode &$node): void {}
}

class CycleSecond extends CycleFirst {}
