<?php

final class Link
{
    public ?Link $next = null;

    public function label(): ?string { return null; }

    public function own(): ?string
    {
        return $this?->label();
    }
}

function known(?Link $given, array $pool, bool $flag)
{
    $made = new Link();
    if ($flag) {
        $maybe = new Link();
    }
    foreach ($pool as $each) {
        $seen = $each?->label();
    }
    return [
        $given?->label(),
        $made?->label(),
        ($made)?->label(),
        $made->next?->label(),
        $pool['k']?->label(),
        $absent?->next->label(),
        // `??` keeps an unset receiver silent: not equivalent alone
        $absent?->label() ?? null,
        $maybe?->label() ?? null,
        $$flag?->label() ?? null,
        $seen ?? null,
    ];
}

$handler = fn () => $outer?->label() ?? null;
$top = $global?->label() ?? null;
