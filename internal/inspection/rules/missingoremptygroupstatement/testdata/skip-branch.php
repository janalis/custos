<?php
function render(array $e, string $g): array
{
    if (!isset($e[$g])) {
        // Intentionally empty: rendered in place.
    }
    elseif (!empty($e['details'])) {
        // Intentionally empty: rendered inside the group.
    }
    elseif ($e[$g]) {
        $e['printed'] = true;
    }

    if ($g === '') {
        // nothing to do
    } else {
        $e[] = $g;
    }

    <weak_warning descr="This construct has an empty body block.">if</weak_warning> ($g === 'x') {
    } else {
        $e[] = 1;
    }

    if ($g === 'y') {
        $e[] = 2;
    } <weak_warning descr="This construct has an empty body block.">elseif</weak_warning> ($g === 'z') {
        // last of the chain: nothing follows
    }

    <weak_warning descr="This construct has an empty body block.">if</weak_warning> ($g === 'w') {
        // alone
    }

    return $e;
}
