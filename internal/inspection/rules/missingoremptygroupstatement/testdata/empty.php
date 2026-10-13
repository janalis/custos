<?php
<weak_warning descr="This construct has an empty body block.">foreach</weak_warning> ($queue as $job) {
    // nothing yet
}
<weak_warning descr="This construct has an empty body block.">while</weak_warning> (poll()) {}
if ($ready) { start(); } <weak_warning descr="This construct has an empty body block.">else</weak_warning> { }
if ($ready) { ; }
if ($ready) /* why */ { start(); } else /* note */ if ($x) { go(); }
if ($ready):
    start();
else:
    stop();
endif;
switch ($x) {}
try {} catch (Exception $e) {}
