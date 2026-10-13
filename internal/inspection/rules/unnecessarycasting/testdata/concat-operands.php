<?php
final class Person
{
    public function __toString(): string { return 'a'; }
}

final class Node
{
    public ?string $slug = null;
    public string $name = '';
    public function label(): string { return ''; }
}

/**
 * @param non-empty-string $title
 * @param positive-int $page
 * @param numeric-string $num
 * @param array<string, mixed> $filters
 * @param scalar $sc
 * @param array-key $key
 */
function concat(string $title, int $page, string $num, array $filters, Person $person, ?Node $node, Node $n, $sc, mixed $m, float $ratio, bool $flag, $key, string $f)
{
    return [
        '%' . (string) $filters['q'] . '%',
        ':' . (string) $node?->slug,
        ':' . (string) $node?->label(),
        ':' . (string) $n->slug,
        (string) preg_replace('/a/', 'b', $title) . '!',
        'at ' . (string) strpos($f, '.'),
        (string) $person . ' / ',
        (string) new Person() . '!',
        (string) $sc . '',
        (string) $m . '',
        (string) $flag . '',
        (string) $undefined . '',
        <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $title . ' - ' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $page,
        'n=' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $num,
        'r=' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> ($ratio * 2),
        'k=' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $key,
        'n=' . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $n->name . <weak_warning descr="Concatenation converts to string anyway; remove the cast.">(string)</weak_warning> $n->label(),
    ];
}
