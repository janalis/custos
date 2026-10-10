<?php
$data = <warning descr="Preserve large JSON integers as strings.">json_decode('{"id":9223372036854775808}', true)</warning>;
