<?php
$s=gzdecode($bytes); <warning descr="Check decompression before consuming data.">strlen($s)</warning>;
