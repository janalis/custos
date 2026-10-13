<?php function big($a) { <warning descr="Return the condition directly: 'return $a > 1'.">if</warning> ($a > 1) { return true; } return false ?>
<?php } function spaced($a) { <warning descr="Return the condition directly: 'return $a <= 1'.">if</warning> ($a > 1) { return false; } return true   ?> <?php }
function alt($a) { ?>
<?php <warning descr="Return the condition directly: 'return $a > 1'.">if</warning> ($a > 1): return true; else: return false; endif ?>
<?php } function inner($a) { <warning descr="Return the condition directly: 'return $a <= 1'.">if</warning> ($a > 1) { return false; } else { return true ?>
<?php } } function commented($a) { <warning descr="Return the condition directly: 'return $a > 1'.">if</warning> ($a > 1) { return true; } /* ?> */ return false; }
