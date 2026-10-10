<?php
// @custos-ignore FailedClosureBindingInvoked
$f = static function () {}; $bound = $f->bindTo(new stdClass()); $bound();
