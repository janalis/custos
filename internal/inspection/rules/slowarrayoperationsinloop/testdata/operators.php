<?php
function scan(array $items, bool $ok, string $s): void
{
    for ($i = 0; $ok && count($items); $i++) {}
    for ($i = 0; $i + strlen($s); $i++) {}
    for ($i = 0; $i <=> count($items); $i++) {}
    for ($i = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$i != count($items)</error>; $i++) {}
    for ($j = 0; <error descr="'strlen(...)' is re-evaluated on every iteration; compute it once before the loop.">$j <= strlen($s)</error>; $j++) {}
}
