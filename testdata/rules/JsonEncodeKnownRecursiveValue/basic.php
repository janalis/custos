<?php
$a = []; $a['loop'] = &$a; <error descr="Remove the cycle before JSON encoding.">json_encode($a, JSON_THROW_ON_ERROR)</error>;
