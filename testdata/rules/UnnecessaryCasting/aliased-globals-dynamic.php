<?php
$dynamic = '';

function loadAny($key) { $GLOBALS[$key] = 4; }

loadAny('dynamic');

return (string) $dynamic;
