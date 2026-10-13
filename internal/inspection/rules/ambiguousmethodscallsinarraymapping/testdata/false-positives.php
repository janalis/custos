<?php
foreach ($users as $user) {
    $byName[strtolower($user)] = true;
    $plain['x'] = strtolower($user);
    $other[md5($user)] = md5($guest);
    $obj[$a->id()] = $b->id();
    $this->prop = f($user);
    if ($user) {
        $skip[md5($user)] = md5($user);
    }
}

foreach ($users as $user)
    $brace[md5($user)] = md5($user);

while ($row = next($rows)) {
    $ids[key($row)] = key($row);
}

foreach ($items as $item) {
    $lazy[$item->getId()] = fn () => $loader->load($item->getId());
    $lazy[$item->getId()] = static fn () => $item->getId();
    $lazy[strtolower($item->name)] = function () use ($item) { return strtolower($item->name); };
    $lazy[$item->getId()] = new class { function m($item) { return $item->getId(); } };
    $lazy[(function () use ($item) { return md5($item); })()] = 'x' . md5($item);
}
