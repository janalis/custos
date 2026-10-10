<?php
$f = static function () {}; $bound = $f->bindTo(null); if ($bound !== null) { $bound(); }
