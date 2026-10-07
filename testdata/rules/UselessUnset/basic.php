<?php
function consume(array &$queue, $item, ...$rest)
{
    static $seen;
    global $registry;

    unset($queue['head']);
    unset($seen, $registry);
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($item);</weak_warning>
    unset($seen, <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">$rest</weak_warning>);
    unset(
        <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">$queue</weak_warning>,
        <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">$item</weak_warning>
    );
    if ($item) {
        <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($rest);</weak_warning>
    }
}

class Pool
{
    public function release($handle)
    {
        $handle = null;
        <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($handle);</weak_warning>
    }
}

$cb = function ($x) {
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($x);</weak_warning>
};
