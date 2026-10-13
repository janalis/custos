<?php
class ExitGuardNode {
    public $value;
    public function next(): ?ExitGuardNode { return null; }
    public function guarded(?ExitGuardNode $node) {
        if (null === $node?->next()) {
            $node = $this->next();
            ((exit(...))(1));
        }
        $node->value = 1;
    }
    public function callableOnly(?ExitGuardNode $node) {
        if (null === $node?->next()) {
            $node = $this->next();
            (exit(...))(...);
        }
        <warning descr="Possible null dereference.">$node</warning>->value = 1;
    }
}
