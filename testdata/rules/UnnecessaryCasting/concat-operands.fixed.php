<?php
final class Author
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
function concat(string $title, int $page, string $num, array $filters, Author $author, ?Node $node, Node $n, $sc, mixed $m, float $ratio, bool $flag, $key, string $f)
{
    return [
        '%' . (string) $filters['q'] . '%',
        ':' . (string) $node?->slug,
        ':' . (string) $node?->label(),
        ':' . (string) $n->slug,
        (string) preg_replace('/a/', 'b', $title) . '!',
        'at ' . (string) filemtime($f),
        (string) $author . ' / ',
        (string) new Author() . '!',
        (string) $sc . '',
        (string) $m . '',
        (string) $flag . '',
        (string) $undefined . '',
        $title . ' - ' . $page,
        'n=' . $num,
        'r=' . ($ratio * 2),
        'k=' . $key,
        'n=' . $n->name . $n->label(),
    ];
}
