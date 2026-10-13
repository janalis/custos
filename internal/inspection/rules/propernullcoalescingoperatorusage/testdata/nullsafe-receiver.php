<?php

final class Link
{
    public ?Link $next = null;

    public function label(): ?string { return null; }

    public function own(): ?string
    {
        return <weak_warning descr="'$this?->label()' alone is equivalent; drop the '?? null' fallback.">$this?->label() ?? null</weak_warning>;
    }
}

function known(?Link $given, array $pool, bool $flag)
{
    $made = new Link();
    if ($flag) {
        $maybe = new Link();
    }
    foreach ($pool as $each) {
        $seen = <weak_warning descr="'$each?->label()' alone is equivalent; drop the '?? null' fallback.">$each?->label() ?? null</weak_warning>;
    }
    return [
        <weak_warning descr="'$given?->label()' alone is equivalent; drop the '?? null' fallback.">$given?->label() ?? null</weak_warning>,
        <weak_warning descr="'$made?->label()' alone is equivalent; drop the '?? null' fallback.">$made?->label() ?? null</weak_warning>,
        <weak_warning descr="'($made)?->label()' alone is equivalent; drop the '?? null' fallback.">($made)?->label() ?? null</weak_warning>,
        <weak_warning descr="'$made->next?->label()' alone is equivalent; drop the '?? null' fallback.">$made->next?->label() ?? null</weak_warning>,
        <weak_warning descr="'$pool['k']?->label()' alone is equivalent; drop the '?? null' fallback.">$pool['k']?->label() ?? null</weak_warning>,
        <weak_warning descr="'$absent?->next->label()' alone is equivalent; drop the '?? null' fallback.">$absent?->next->label() ?? null</weak_warning>,
        // `??` keeps an unset receiver silent: not equivalent alone
        $absent?->label() ?? null,
        $maybe?->label() ?? null,
        $$flag?->label() ?? null,
        $seen ?? null,
    ];
}

$handler = fn () => $outer?->label() ?? null;
$top = $global?->label() ?? null;
