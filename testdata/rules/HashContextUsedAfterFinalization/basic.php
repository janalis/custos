<?php
$h = hash_init('sha256'); hash_final($h); <error descr="Create a new hash context after finalization.">hash_update($h, 'next')</error>;
