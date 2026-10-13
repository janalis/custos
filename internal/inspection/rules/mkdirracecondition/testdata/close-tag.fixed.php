<?php function prepare($d, $c) { ?>
<b><?php if (!mkdir($d) && !is_dir($d)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $d)); } ?></b>
<i><?php if ($c) { if (!mkdir($d) && !is_dir($d)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $d)); } } ?></i>
<?php }
