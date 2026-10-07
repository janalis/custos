<?php
while (<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning>($key, $item) = each($legacy)) {
    use_it($key, $item);
}
$copy = <weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning>($x, $y) = $coords;
for (<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning>($i) = $start; $i < 10; $i++) {
}
if (<weak_warning descr="Use short destructuring syntax '[...] = ...'.">list</weak_warning>($a, list($b, $c)) = pair()) {
}
