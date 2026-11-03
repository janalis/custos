<?php
class Plain implements Labeled {
    public function label(): ?string { return null; }
    public function name() { return 'plain'; }
}
class Leaf extends Node {
    /** @return null */
    public function setName($name) { }
    public function maybe(bool $b) { return null; }
    public function kept() { return $this; }
}
