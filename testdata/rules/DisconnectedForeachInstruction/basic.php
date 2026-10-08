<?php
/* @var array $orders */
$type = \ArrayObject::class;
foreach ($orders as $key => $order) {
    <weak_warning descr="Create the object once before the loop and clone it here.">$box = new $type();</weak_warning>
    <weak_warning descr="Create the object once before the loop and clone it here.">$node = (new \DOMDocument())->createElement('row');</weak_warning>

    <weak_warning descr="Statement does not depend on the loop; move it out.">syslog(LOG_INFO, $banner);</weak_warning>
    <weak_warning descr="Statement does not depend on the loop; move it out.">while</weak_warning> ($pending > 0) {
        usleep($pending);
    }
    <weak_warning descr="Statement does not depend on the loop; move it out.">if</weak_warning> ($debugMode) {
        dump($settings);
    }
    <weak_warning descr="Statement does not depend on the loop; move it out.">switch</weak_warning> ($mode) {
        default:
    }
    <weak_warning descr="Statement does not depend on the loop; move it out.">try</weak_warning> {
        warmup();
    } catch (\RuntimeException $problem) {
        report($problem);
    }

    $copy = clone $template;            // clone: never reported
    echo '<br>';                        // E4: no variables
    if ($stop) { return; }              // E3
    $total += $order;                   // E2 + connected
    $sink[] = 'seen';                   // accumulate
    $logger->push($order);              // D8a: $logger becomes modified
    $logger->flush();                   // connected through $logger
    preg_match('/x/', $order, $hits);   // D8b: $hits modified
    print_r($hits);                     // connected
    $meta->touch(compact('key'));       // D11: depends on $key
}

foreach ($groups as $group) {
    foreach ($group as $member) {
        echo $group;                    // D4: outer loop variable
    }
}

foreach ($rows as $row) {
    ?><td><?= $title ?></td><?php       // E5: inline HTML, loop skipped
}

foreach ($events as $event) {
    <weak_warning descr="Statement does not depend on the loop; move it out.">Registry::$aliases = $aliases;</weak_warning>
    handle($event);
}
