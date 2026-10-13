<?php
// @custos-ignore WeakMapValueRetainsKey
$m = new WeakMap(); $o = new stdClass(); $m[$o] = $o;
