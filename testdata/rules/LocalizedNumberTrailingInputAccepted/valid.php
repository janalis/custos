<?php
$f = new NumberFormatter("en_US", NumberFormatter::DECIMAL); $s="37tail"; $n=$f->parse($s,NumberFormatter::TYPE_DOUBLE,$end); if ($n !== false && $end === strlen($s)) { echo $n; }
