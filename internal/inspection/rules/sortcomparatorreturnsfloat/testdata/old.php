<?php
// Before PHP 7.0 there is no <=> to offer as a quick-fix.
<warning descr="Compare float values without subtracting them.">usort($xs, function ($a, $b) { return (float) $a - (float) $b; })</warning>;
