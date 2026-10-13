<?php
function f(float $value) { if(<warning descr="Use is_nan to test for NaN.">$value===NAN</warning>) { echo "bad"; } }
