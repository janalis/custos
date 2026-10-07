<?php
// __serialize() must return array and __unserialize() void: an always
// throwing body would otherwise get ': void', which PHP rejects.
final class Locked {
    public function __serialize() { throw new \LogicException('locked'); }
    public function __unserialize(array $data) { throw new \LogicException('locked'); }
}
