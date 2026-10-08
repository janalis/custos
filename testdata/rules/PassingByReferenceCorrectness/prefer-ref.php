<?php

namespace Shop\Sorting {
    function array_multisort(array &$rows, int $flags = 0): bool { return true; }
    function extract(array &$vars): int { return 0; }

    function local(array $rows): void
    {
        array_multisort(<warning descr="Pass a variable here: this parameter is taken by reference.">array_values($rows)</warning>);
        \array_multisort(array_values($rows), SORT_ASC, array_keys($rows), SORT_DESC, $rows);
        extract(<warning descr="Pass a variable here: this parameter is taken by reference.">array_merge($rows, [])</warning>);
        \extract(array_merge($rows, []));
    }
}

namespace {
    function rank(array $scores): array
    {
        array_multisort(array_values($scores), SORT_DESC, array_keys($scores), SORT_ASC, $scores);
        return $scores;
    }

    function render(array $vars, array $defaults): void
    {
        extract(array_merge($defaults, $vars));
        extract(defaults(), EXTR_SKIP);
    }

    function defaults(): array { return []; }
}
