<?php $greeting = 'hi'; ?>
<p>
<?php foreach ($users as $u) : ?>
  <b><?= $u ?></b>
<?php endforeach ?>
<?php
  while (next_row()) { ?><i>row</i><?php } ?>
</p>
<?php $done = true; ?>
<?php	$tab = 1 ?>
