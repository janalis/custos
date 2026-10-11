<?php
function asserted($d) { $fs = scandir($d); assert($fs !== false); foreach (<warning descr="Check directory enumeration before iterating.">$fs</warning> as $f) { echo $f; } }
function conjunctionExit($d, $c) { $fs = scandir($d); if (!$fs && $c) { return; } foreach (<warning descr="Check directory enumeration before iterating.">$fs</warning> as $f) { echo $f; } }
function disjunctionBranch($d, $c) { $fs = scandir($d); if ($fs || $c) { foreach (<warning descr="Check directory enumeration before iterating.">$fs</warning> as $f) { echo $f; } } }
function wrongBranch($d) { $fs = scandir($d); if ($fs) { echo 1; } else { foreach (<warning descr="Check directory enumeration before iterating.">$fs</warning> as $f) { echo $f; } } }
function emptyBranch($d) { $fs = scandir($d); if (empty($fs)) { foreach (<warning descr="Check directory enumeration before iterating.">$fs</warning> as $f) { echo $f; } } }
function otherLocal($d, $o) { $fs = scandir($d); if (!$o || !is_array($o) || $o == false || empty($o)) { return; } foreach (<warning descr="Check directory enumeration before iterating.">$fs</warning> as $f) { echo $f; } }
function unrelatedTests($d, $o) { $fs = scandir($d); if (-$o) { return; } if ($fs > 1) { return; } if ($fs == null) { return; } if (is_string($fs)) { return; } echo 1; foreach (<warning descr="Check directory enumeration before iterating.">$fs</warning> as $f) { echo $f; } }
function notTerminating($d) { $fs = scandir($d); if (!$fs) { echo 0; } foreach (<warning descr="Check directory enumeration before iterating.">$fs</warning> as $f) { echo $f; } }
function elseifExit($d, $c) { $fs = scandir($d); if ($fs) { echo 1; } elseif ($c) { echo 2; } else { return; } foreach (<warning descr="Check directory enumeration before iterating.">$fs</warning> as $f) { echo $f; } }
function direct($d, $c) { if ($c) { foreach (<warning descr="Check directory enumeration before iterating.">scandir($d)</warning> as $f) { echo $f; } } }
function rewritten($d, $ds) { $fs = scandir($d); if (!$fs) { return; } while ($ds) { foreach (<warning descr="Check directory enumeration before iterating.">$fs</warning> as $f) { echo $f; } $fs = scandir(array_pop($ds)); } }
