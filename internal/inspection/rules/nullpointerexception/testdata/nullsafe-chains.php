<?php
final class SafeBranch {
    public function next(): SafeBranch { return $this; }
    public function maybe(): ?SafeBranch { return null; }
    public function label(): string { return 'branch'; }
    /** @return ?SafeBranch */
    public function documented() { return null; }
    public function inferred() { return null; }
}

function branchLabels(?SafeBranch $root) {
    $root?->next()->next()->label();
    $root?->next()->next()->next()->label();
    $root?->next()->maybe()<warning descr="Possible null dereference.">-></warning>label();
    $root?->next()->documented()<warning descr="Possible null dereference.">-></warning>label();
    $root?->next()->inferred()<warning descr="Possible null dereference.">-></warning>label();
    null?->next()->next()->label();

    $saved = $root?->next()->next();
    <warning descr="Possible null dereference.">$saved</warning>->label();
}
