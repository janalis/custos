<?php
// @custos-ignore NeverFunctionFallsThrough
function halt(bool $stop): never { if ($stop) { exit; } }
