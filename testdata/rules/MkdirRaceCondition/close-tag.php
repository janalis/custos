<?php function prepare($d, $c) { ?>
<b><?php <error descr="mkdir() outcome is ignored; use 'if (!mkdir($d) && !is_dir(...)) { ... }'.">mkdir($d) ?></error></b>
<i><?php if ($c) <error descr="mkdir() outcome is ignored; use 'if (!mkdir($d) && !is_dir(...)) { ... }'.">mkdir($d) ?></error></i>
<?php }
