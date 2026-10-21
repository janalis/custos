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

// A nullsafe read equal to a value that is never null proves the base is not
// null, also for a local whose own type is not declared.
enum RowKind { case Header; case Line; }
final class MergeRow { public function __construct(public RowKind $kind, public ?int $region) {} }
final class MergePane {
    /** @var list<MergeRow> */
    private array $rows = [];
    public function press(int $i): void {
        $row = $this->rows[$i] ?? null;
        if (RowKind::Header === $row?->kind && null !== $row->region) {
            $this->accept($row->region);
        }
        if (RowKind::Line !== $row?->kind) {
            return;
        }
        $this->accept($row->region);
        $other = $this->rows[$i + 1] ?? null;
        if ($other?->region === null) {
            <warning descr="Possible null dereference.">$other</warning>->kind;
        }
    }
    private function accept(?int $region): void {}
}
