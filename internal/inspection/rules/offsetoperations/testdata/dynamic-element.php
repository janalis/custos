<?php
function totals(bool $byItem, string $key, int $value): array
{
    $days = $byItem ? [] : ['total' => 0, 'label' => null];
    if ($byItem) {
        $days[$key] = ['total' => 0];
        $days[$key]['total'] += $value;
        $days[($key)]['total'] += $value;
    } else {
        $days['total'] += $value;
        <error descr="'$days['total']' does not support offset access (types: int).">$days['total']['x']</error>;
    }
    return $days;
}

function flags(string $key): void
{
    $marks = ['a' => 1, 'b' => 2];
    <error descr="'$marks[$key]' does not support offset access (types: int).">$marks[$key]['x']</error>;
}
