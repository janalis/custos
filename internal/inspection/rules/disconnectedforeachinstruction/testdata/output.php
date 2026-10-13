<?php
// Output and by-reference method arguments happen on every iteration.
final class StringHelper
{
    public static function increment(string &$str): void
    {
        $str++;
    }
}

final class Counter
{
    public function bump(int &$n): void
    {
        $n++;
    }

    public function __construct(?array &$log = null)
    {
        $log[] = 'created';
    }
}

function fill(array $values, string $column, Counter $c, int $n, array $log): array
{
    $out = [];
    foreach ($values as $value) {
        $out[$column] = $value;
        StringHelper::increment($column);
    }
    foreach ($values as $value) {
        $out[$n] = $value;
        $c->bump($n);
    }
    foreach ($values as $value) {
        $out[] = [$value, $log];
        new Counter($log);
    }
    return $out;
}

function progress(array $rows, bool $display, $page): void
{
    foreach ($rows as $row) {
        handle($row);
        if ($display) {
            echo '.';
        }
        echo $page->boxStart();
        print $page->separator();
        printf('%s', $page->title());
        <weak_warning descr="Statement does not depend on the loop; move it out.">store($page->footer());</weak_warning>
    }
}

function handle($r) {}
function store($v) {}

function dynamicMethod(array $rows, $o, string $m, $acc)
{
    foreach ($rows as $row) {
        handle($row);
        $o->$m($acc);
    }
}
