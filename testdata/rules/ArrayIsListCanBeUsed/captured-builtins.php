<?php
namespace Lists {
    function array_is_list(array $a): bool { return false; }

    function check(array $rows): array
    {
        return [
            <weak_warning descr="Replace with '\array_is_list($rows)'.">array_values($rows) === $rows</weak_warning>,
            <weak_warning descr="Replace with '!\array_is_list($rows)'.">\array_values($rows) !== $rows</weak_warning>,
        ];
    }
}

namespace Imported {
    use function Lists\array_is_list;

    function check(array $rows): bool
    {
        return <weak_warning descr="Replace with '\array_is_list($rows)'.">$rows === array_values($rows)</weak_warning>;
    }
}

namespace Plain {
    function check(array $rows): bool
    {
        return <weak_warning descr="Replace with 'array_is_list($rows)'.">$rows === array_values($rows)</weak_warning>;
    }
}
