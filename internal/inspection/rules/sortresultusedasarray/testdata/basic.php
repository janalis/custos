<?php
$sorted = sort($items); <warning descr="Use the sorted input array instead of the success flag.">foreach ($sorted as $item) {}</warning>
