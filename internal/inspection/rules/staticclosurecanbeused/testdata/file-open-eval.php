<?php
$d = function () { return 1; };
eval('$d->bindTo(new stdClass);');
