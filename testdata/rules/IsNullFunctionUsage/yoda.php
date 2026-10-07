<?php

$x = <weak_warning descr="Replace with 'null !== $token'.">true !== is_null($token)</weak_warning>;
$y = <weak_warning descr="Replace with 'null === ($m ?: $n)'.">is_null($m ?: $n)</weak_warning>;
