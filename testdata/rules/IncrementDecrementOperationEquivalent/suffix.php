<?php

<weak_warning descr="Use '$page++' instead.">$page += 1</weak_warning>;
<weak_warning descr="Use '$this->left--' instead.">$this->left -= 1</weak_warning>;
<weak_warning descr="Use '$total++' instead.">$total = 1 + $total</weak_warning>;
<weak_warning descr="Use '$total++' instead.">$total = $total + 1</weak_warning>;
<weak_warning descr="Use 'self::$depth--' instead.">self::$depth = self::$depth - 1</weak_warning>;
$total = 1 - $total;
$total = $total + 1.0;

$slots = [0, 0];
<weak_warning descr="Use '$slots[1]++' instead.">$slots[1] += 1</weak_warning>;

while (true) { <weak_warning descr="Use '$n--' instead.">$n = $n - 1</weak_warning>; }
for ($i = 0; $i < 3; <weak_warning descr="Use '$i++' instead.">$i += 1</weak_warning>) {}
