<?php
class ChainLabel {
    public string $name = 'chain';
    private string $hidden = 'chain';
    public function next(): ChainLabel { return $this; }
    public function label(): string { return 'chain'; }
    public function hidden(?ChainLabel $other): string { return (string) $other?->hidden; }
}

function chainLabels(?ChainLabel $maybe, ChainLabel $present) {
    $names = [$maybe?->name];
    return [
        (string) $names[0],
        'nullable=' . (string) $maybe?->next()->label(),
        'nested=' . (string) $maybe?->next()->next()->label(),
        'present=' . $present->next()->label(),
    ];
}
