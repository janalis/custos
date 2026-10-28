<?php
function flush_all($items, $q)
{
    while (@ob_end_clean()) { }
    while ($row = $q->fetch()) {}
    do {} while (++$items < 10);
    while (Pool::drain()) {}
    while (new Pool() && $items--) {}
    <weak_warning descr="This construct has an empty body block.">while</weak_warning> ($items) {}
    <weak_warning descr="This construct has an empty body block.">while</weak_warning> (function () { return next(); }) {}
    <weak_warning descr="This construct has an empty body block.">do</weak_warning> {} while ($items > 0);
    <weak_warning descr="This construct has an empty body block.">for</weak_warning> (; next($items);) {}
}
