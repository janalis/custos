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

// The value is used: only the prefix form keeps it.
function next_id(array $ids) {
    $b = <weak_warning descr="Use '++$a' instead.">$a += 1</weak_warning>;
    while ((<weak_warning descr="Use '++$i' instead.">$i = $i + 1</weak_warning>) < 3) {}
    for (<weak_warning descr="Use '$j++' instead.">$j += 1</weak_warning>, $k = 0; <weak_warning descr="Use '++$k' instead.">$k += 1</weak_warning>; (<weak_warning descr="Use '$k--' instead.">$k -= 1</weak_warning>)) {}
    (<weak_warning descr="Use '$b++' instead.">$b += 1</weak_warning>);
    return <weak_warning descr="Use '--$ids[0]' instead.">$ids[0] -= 1</weak_warning>;
}
