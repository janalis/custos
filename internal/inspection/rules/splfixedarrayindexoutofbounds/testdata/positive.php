<?php
$a = new SplFixedArray(3); <error descr="Use an index within the fixed array size.">$a[3]</error> = "outside";
