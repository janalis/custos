<?php
// The unset changes what later code of the function sees.
function headers(string $action, ?int $userid = null): array
{
    $headers = [];
    if ($action === 'add' && $userid) {
        unset($userid);
    }
    if (!empty($userid)) {
        $headers['user'] = $userid;
    }
    return $headers;
}

function merge(array $config, array $overrides)
{
    foreach ($overrides as $key => $value) {
        if ($key === 'reset') {
            unset($config);
            continue;
        }
        $config[$key] = $value;
    }
    return $config ?? [];
}

// Still pointless: only overwritten afterwards, or never touched again.
function overwritten($a, $b)
{
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($a);</weak_warning>
    $a = $b;
    return $a;
}

function looped(array $rows, $tmp)
{
    foreach ($rows as $r) {
        <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($tmp);</weak_warning>
        echo $r;
    }
}

function reassignedFirst($p, $c)
{
    $p = 1;
    if ($c) {
        unset($p);
    }
    return isset($p);
}
