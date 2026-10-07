<ul>
<li><?= $already ?></li>
<li><?php echo $a; echo $b; ?></li>
<li><?php /* hint */ echo $c ?></li>
<li><?php echo $d; /* hint */ ?></li>
<li><?php $r = print $x ?></li>
<li><?php if ($x) echo $y; ?></li>
<li><?php foreach ($rows as $r): ?><?php endforeach ?></li>
</ul>
<?php echo $last;
