<?php
// @custos-ignore SplObjectStorageIndirectArrayMutation
$s = new SplObjectStorage(); $o = new stdClass(); $s[$o] = ["count" => 1]; $s[$o]["count"]++;
