<?php
// Before PHP 7.0 there is no <=> to offer as a quick-fix.
<warning descr="Return a three-way integer comparison result.">usort($xs, function ($a, $b) { return $a > $b; })</warning>;
