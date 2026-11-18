<?php
$h = hash_init('sha256'); hash_update($h, 'next'); hash_final($h);
