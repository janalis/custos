<?php
function tally(int $count, string $label, array $grid, $buf) {
    <weak_warning descr="Use the compound form '$count += 5'.">$count = $count + 5</weak_warning>;
    <weak_warning descr="Use the compound form '$count -= $step'.">$count = $count - $step</weak_warning>;
    <weak_warning descr="Use the compound form '$count %= 7'.">$count = ($count % 7)</weak_warning>;
    <weak_warning descr="Use the compound form '$count <<= 1'.">$count = $count << 1</weak_warning>;
    <weak_warning descr="Use the compound form '$label .= '-' . $count . '!''.">$label = $label . '-' . $count . '!'</weak_warning>;
    <weak_warning descr="Use the compound form '$count *= 3 * $k'.">$count = $count * 3 * $k</weak_warning>;
    <weak_warning descr="Use the compound form '$count += ($k * 2)'.">$count = $count + ($k * 2)</weak_warning>;
    <weak_warning descr="Use the compound form '$grid[1] |= 4'.">$grid[1] = $grid[1] | 4</weak_warning>;
    <weak_warning descr="Use the compound form '$buf[0] .= 'z''.">$buf[0] = $buf[0] . 'z'</weak_warning>;
    <weak_warning descr="Use the compound form '$this->n ^= $m'.">$this -> n = $this->n ^ $m</weak_warning>;
}
