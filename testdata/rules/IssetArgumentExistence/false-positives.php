<?php
class Billing
{
    public function total($items)
    {
        $cb = function ($rate) use ($items) {
            return [$rate ?? 1, isset($items), $_POST ?? null, isset($this), ($x) ?? 2];
        };
        return $cb;
    }

    public function known()
    {
        echo $label;
        return $label ?? 'none';
    }

    public function looping($rows)
    {
        foreach ($rows as $row) {
            $prev = isset($last) ? $last : null;
            $last = $row;
        }
    }

    public function jumping()
    {
        again:
        $r = isset($tries) ? $tries : 0;
        $tries = 1;
        goto again;
    }

    public function arrow()
    {
        return fn() => $missing ?? 0;
    }

    public function statics()
    {
        static $count;
        return isset($count, $this->x, $a['k']) || isset(self::$p);
    }

    public function viaGlobal()
    {
        global $conf;
        return $conf ?? null;
    }

    public function caught()
    {
        try {
        } catch (\Exception $e) {
        }
        return isset($e);
    }
}
$top = $undefinedAtTop ?? 1;

function assignedEarlier()
{
    $known = 1;
    return isset($known);
}

function connectionString(array $options)
{
    extract($options, EXTR_SKIP);
    return isset($hostname) ? "host=$hostname" : '';
}

function fromQuery(string $query)
{
    parse_str($query);
    return $page ?? 1;
}

function dynamicName(string $key)
{
    $$key = true;
    return empty($enabled);
}

function loadVersion(string $dir)
{
    require $dir . '/version.php';
    return $release ?? 'dev';
}

// Set by a later statement of an outer loop, for the next iteration.
function rows(array $ids): array
{
    $out = [];
    foreach ($ids as $id) {
        foreach ([1, 2] as $k) {
            if (isset($prev) && $prev === $k) {
                $out[] = $id;
            }
        }
        $prev = $id;
    }
    return $out;
}

// Defined for the next iteration by a later statement of the loop (D8).
function byValue($value, $other = null)
{
    return $value;
}

function byReference($source, &$target, &...$rest)
{
    $target = $source;
}

function carriedOver(array $rows, callable $unknown)
{
    foreach ($rows as $row) {
        if (isset($seen)) {
            echo count($seen);
        }
        $seen[] = $row;
    }
    foreach ($rows as $row) {
        $previous = $last ?? null;
        $last['row']['id'] = $row;
    }
    foreach ($rows as $line) {
        if (isset($m)) {
            echo $m[0];
        }
        preg_match('/\d+/', $line, $m);
    }
    foreach ($rows as $group) {
        if (isset($item)) {
            echo $item;
        }
        foreach ($group as $item) {
        }
    }
    foreach ($rows as $group) {
        if (isset($index)) {
            echo $index;
        }
        foreach ($group as $index => $ignored) {
        }
    }
    foreach ($rows as $group) {
        if (isset($head)) {
            echo $head;
        }
        foreach ($group as [$head, $tail]) {
        }
    }
    for ($i = 0; $i < 2; $i++) {
        if (isset($e)) {
            echo $e->getMessage();
        }
        try {
            byValue($i);
        } catch (\LogicException) {
        } catch (\RuntimeException $e) {
        }
    }
    while ($unknown()) {
        if (isset($n)) {
            echo $n;
        }
        $n++;
    }
    do {
        if (isset($buffer)) {
            echo $buffer;
        }
        $buffer .= 'x';
    } while ($unknown());
    foreach ($rows as $row) {
        if (isset($alias)) {
            echo 1;
        }
        $copy = &$alias;
    }
    foreach ($rows as $row) {
        if (isset($shared)) {
            echo 1;
        }
        global $shared;
    }
    foreach ($rows as $row) {
        if (isset($kept)) {
            echo 1;
        }
        static $kept = 1;
    }
    foreach ($rows as $row) {
        if (isset($bound)) {
            echo 1;
        }
        $set = function () use ($row, &$bound) {
            $bound = 1;
        };
        $set();
    }
    foreach ($rows as $row) {
        if (isset($named)) {
            echo 1;
        }
        byReference(target: $named, source: $row);
    }
    foreach ($rows as $row) {
        if (isset($extra)) {
            echo 1;
        }
        byReference($row, $row, $row, $extra);
    }
    foreach ($rows as $row) {
        if (isset($spread)) {
            echo 1;
        }
        byValue(...$rows, other: $spread);
    }
    foreach ($rows as $row) {
        if (isset($opaque)) {
            echo 1;
        }
        $unknown($opaque);
    }
    foreach ($rows as $row) {
        if (isset($built)) {
            echo 1;
        }
        new \Unknown\Collector($built['list']);
    }
    foreach ($rows as $row) {
        if (isset($viaMethod)) {
            echo 1;
        }
        $row->collect($viaMethod);
    }
    foreach ($rows as $row) {
        if (isset($viaStatic)) {
            echo 1;
        }
        \Unknown\Registry::collect($viaStatic);
    }
}

function evaluated(string $code)
{
    eval($code);
    return isset($result);
}
