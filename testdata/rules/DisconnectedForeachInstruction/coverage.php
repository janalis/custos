<?php
class Counter { public static $hits = 0; public static $list = []; }
function cover(array $orders, $cfg, $doc, $other, $o, $s) {
    foreach ($orders as $order) { /* only a comment */ }
    foreach ($orders as $order) {
        $cfg->last->id = $order;               // property write: modifies $cfg
        [$head, $tail] = explode(',', $order); // destructuring targets
        echo $head, $tail;
        ($o)->add($order);                     // receiver through parentheses
        $o->items->add($order);                // receiver through a property
        sscanf($order, '%d %d', $n, $m);       // variadic by-reference parameter
        echo $n, $m;
        Counter::$hits++;                      // no variable: increment, never reported
        Counter::$list[] = 'x';                // no variable: accumulation, never reported
        $flag = 'on';                          // plain assignment, never reported
        $loaded = $doc->load();                // method call other than createElement
        $el = $other->createElement('a');      // createElement on an unknown class
        $rows[0]->seen = $order;               // element write through a property
        $o->keep($cells[0]);                   // method argument is not an element write
        sscanf($order, '%d %d', $p, $q[0]);    // element bound to the variadic by-reference parameter
        echo $q;
        break;
    }
    foreach ($orders as $order) {
        ?><?= $s ?><?php
        handle($order);
    }
}
