<?php

$csv   = <weak_warning descr="Pass the separator as the first argument of implode().">implode($cells, ';')</weak_warning>;
$path  = <weak_warning descr="Pass the separator as the first argument of implode().">\implode( $segments ,  "/" )</weak_warning>;
$line  = <weak_warning descr="Pass the separator as the first argument of implode().">implode(get_rows(), "\n{$eol}")</weak_warning>;
$keys  = <weak_warning descr="Pass the separator as the first argument of implode().">implode(['a', 'b'], ',')</weak_warning>;
