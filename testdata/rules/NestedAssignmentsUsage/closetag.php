<?php <weak_warning descr="Split this chained assignment into separate assignments.">$a = $b = 1</weak_warning> ?>
<p><?php echo $a + $b ?></p>
<?php if ($a) <weak_warning descr="Split this chained assignment into separate assignments.">$c = $d = load()</weak_warning> ?>
<p><?= $c ?></p>
