<?php $b = 1;
$a = 1; ?>
<p><?php echo $a + $b ?></p>
<?php if ($a) {
    $d = load();
    $c = $d;
} ?>
<p><?= $c ?></p>
