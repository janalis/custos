<?php

function f(array $list) {
    <weak_warning descr="Use '++$count' instead.">$count += 1</weak_warning>;
    <weak_warning descr="Use '--$list[0]' instead.">$list[0] = $list[0] - 1</weak_warning>;
}
